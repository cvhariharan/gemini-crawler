package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/cvhariharan/gemini-crawler/gemini"
	"github.com/cvhariharan/gemini-crawler/gemtext"
)

const (
	WORKING_INDEX = "index2"
	INDEX         = "index"
)

type Data struct {
	Path string
	Text string
}

func main() {
	mountPoint := os.Getenv("AWS_EFS_MOUNT")
	if mountPoint == "" {
		log.Fatal("mount point not set. AWS_EFS_MOUNT empty")
	}

	// Create a new index at mountpoint/index2
	indexPath := filepath.Join(mountPoint, WORKING_INDEX)
	fmt.Println("Index path -", indexPath)

	// Setup seed URLs
	q := NewQueue()
	q.Enqueue("gemini://gemini.circumlunar.space/")

	removeIfExists(indexPath)

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New(indexPath, mapping)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	createIndexer(q, index)
	end := time.Now()

	// Rename new index2 to index
	removeIfExists(filepath.Join(mountPoint, INDEX))
	err = os.Rename(WORKING_INDEX, INDEX)
	if err != nil {
		log.Println(err)
	}

	fmt.Printf("Indexing complete in %f minutes\n", end.Sub(start).Minutes())
}

func createIndexer(q *Queue, index bleve.Index) {
	client := gemini.NewClient(gemini.ClientOptions{Insecure: true})
	for q.Q.Len() != 0 {
		path := q.Dequeue()
		fmt.Println(path)
		resp, err := client.Fetch(path)
		if err != nil {
			log.Println(err)
			continue
		}

		if resp.Meta == "text/gemini" {
			txt, _ := ioutil.ReadAll(resp.Body)
			links, _ := gemtext.GetLinks(string(txt), path)
			for _, v := range links {
				if !q.IsAdded(v) {
					q.Enqueue(v)
				}
			}

			index.Index(path, Data{Path: path, Text: string(txt)})
		}
	}
}

func removeIfExists(src string) {
	if _, err := os.Stat(src); err == nil {
		err = os.RemoveAll(src)
		if err != nil {
			log.Fatal(err)
		}
	}
}
