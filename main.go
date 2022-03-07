package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"sync"

	"github.com/blevesearch/bleve"
	"github.com/cvhariharan/gemini-crawler/gemini"
	"github.com/cvhariharan/gemini-crawler/gemtext"
)

type Data struct {
	Path string
	Text string
}

func main() {
	var wg sync.WaitGroup
	q := NewQueue()
	q.Enqueue("gemini://gemini.circumlunar.space/")

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("gemini.bleve", mapping)
	if err != nil {
		log.Fatal(err)
	}

	wg.Add(1)
	go func(q *Queue) {
		client := gemini.NewClient(gemini.ClientOptions{Insecure: true})
		for q.Q.Len() != 0 {
			path := q.Dequeue()
			fmt.Println(path)
			resp, err := client.Fetch(path)
			if err != nil {
				log.Println(err)
				continue
			}
			txt, _ := ioutil.ReadAll(resp.Body)
			g, _ := gemtext.Parse(string(txt), path)
			for _, v := range g.Links {
				if !q.IsAdded(v) {
					q.Enqueue(v)
				}
			}

			index.Index(path, Data{Path: path, Text: string(txt)})
		}
		wg.Done()
	}(q)

	wg.Wait()
	q.PrintAll()
}
