package main

import (
	"container/list"
	"io/ioutil"

	"github.com/cvhariharan/gemini-crawler/gemini"
	"github.com/cvhariharan/gemini-crawler/gemtext"
)

func main() {
	client := gemini.NewClient(gemini.ClientOptions{Insecure: true})
	resp, _ := client.Fetch("gemini://gemini.circumlunar.space/")
	txt, _ := ioutil.ReadAll(resp.Body)

	q := Queue{Q: list.New()}
	g, _ := gemtext.Parse(string(txt), "gemini://gemini.circumlunar.space/")
	for _, v := range g.Links {
		q.Enqueue(v)
	}
	q.PrintAll()
}
