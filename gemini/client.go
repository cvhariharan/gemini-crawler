package gemini

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
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

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}, "tcp", url, &config)
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
	scanner := bufio.NewScanner(req.Conn)
	ok := scanner.Scan()
	if ok {
		header := scanner.Text()
		fields := strings.Fields(header)

		if len(fields) > 1 {
			status, err := strconv.Atoi(fields[0])
			if err != nil {
				log.Println(err)
				return nil, err
			}
			meta := fields[1]

			return &Response{
				Body:   req.Conn,
				Status: status,
				Meta:   meta,
			}, nil
		}
	}

	return nil, errors.New("header not found")
}
