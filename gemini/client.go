package gemini

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"time"
)

type Response struct {
	Status int
	Meta   string
	Body   io.ReadCloser
}

func (r *Response) Close() {
	r.Body.Close()
}

type ClientOptions struct {
	ConnectTimeout time.Duration
	Insecure       bool
}

type Request struct {
	Conn *tls.Conn
}

type Client struct {
	Options ClientOptions
}

func getHostURL(uri string) (string, error) {
	parsed, err := url.Parse(uri)
	host := parsed.Host
	if parsed.Port() == "" {
		host = net.JoinHostPort(host, "1965")
	}
	return host, err
}

func NewClient(options ClientOptions) *Client {
	return &Client{
		Options: options,
	}
}

func (c *Client) Connect(uri string) (*Request, error) {
	var config tls.Config
	config.InsecureSkipVerify = c.Options.Insecure

	url, err := getHostURL(uri)
	if err != nil {
		return nil, err
	}
	conn, err := tls.Dial("tcp", url, &config)
	if err != nil {
		return nil, err
	}

	return &Request{conn}, nil
}

func (c *Client) Fetch(path string) (*Response, error) {
	req, err := c.Connect(path)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	fmt.Fprintf(req.Conn, "%s\r\n", path)
	// message, err := ioutil.ReadAll(req.Conn)
	// if err != nil {
	// 	log.Println(err)
	// }
	// fmt.Println(string(message))
	return &Response{
		Body: req.Conn,
	}, nil
}
