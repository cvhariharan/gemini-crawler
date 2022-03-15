package main

import (
	"bufio"
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
	SEED_FILE     = "seeds.txt"
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
	seeds := []string{"gemini://gemini.circumlunar.space/"}
	if _, err := os.Stat(SEED_FILE); err == nil {
		file, err := os.Open(SEED_FILE)
		if err != nil {
			log.Fatal("could not read seeds from file, delete or replace the file")
		}
		defer file.Close()

		seeds = []string{}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			seeds = append(seeds, scanner.Text())
		}
	}

	q := NewQueue()
	for _, seed := range seeds {
		q.Enqueue(seed)
	}

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
	err = os.Rename(filepath.Join(mountPoint, WORKING_INDEX), filepath.Join(mountPoint, INDEX))
	if err != nil {
		log.Println(err)
	}

	fmt.Printf("Indexing complete in %f minutes\n", end.Sub(start).Minutes())
}

func createIndexer(q *Queue, index bleve.Index) {
	client := gemini.NewClient(gemini.ClientOptions{Insecure: true})
	for q.Q.Len() != 0 {
		path := q.Dequeue()
		log.Println(path)
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
