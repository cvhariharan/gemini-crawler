package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/cvhariharan/gemini-crawler/gemini"
	"github.com/cvhariharan/gemini-crawler/gemtext"
)

const (
	WORKING_INDEX = "index2"
	INDEX         = "index"
	SEED_FILE     = "seeds.txt"
	WORKERS       = 5
)

var LinkLock sync.Mutex

type Data struct {
	Path string
	Text string
}

func main() {
	var wg sync.WaitGroup
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

	c := make(chan string, 100)
	indexChan := make(chan Data, 100)
	q := NewQueue()

	removeIfExists(indexPath)

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New(indexPath, mapping)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	go indexer(indexChan, index, &wg)
	for i := 0; i < WORKERS; i++ {
		wg.Add(1)
		go createCrawler(c, indexChan, q, &wg)
	}

	for _, v := range seeds {
		c <- v
	}

	wg.Wait()
	close(indexChan)

	// Rename new index2 to index
	removeIfExists(filepath.Join(mountPoint, INDEX))
	err = os.Rename(filepath.Join(mountPoint, WORKING_INDEX), filepath.Join(mountPoint, INDEX))
	if err != nil {
		log.Println(err)
	}

	end := time.Now()

	fmt.Printf("Indexing complete in %f minutes\n", end.Sub(start).Minutes())
}

func createCrawler(c chan string, indexChan chan Data, q *Queue, wg *sync.WaitGroup) {
	client := gemini.NewClient(gemini.ClientOptions{Insecure: true})
	for path := range c {
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
				go func(v string) {
					LinkLock.Lock()
					defer LinkLock.Unlock()
					if !q.IsAdded(v) {
						q.Visit(v)
						c <- v
					}
				}(v)
			}

			go func() {
				indexChan <- Data{Path: path, Text: string(txt)}
			}()
		}
	}
	wg.Done()
}

func indexer(indexChan chan Data, index bleve.Index, wg *sync.WaitGroup) {
	for v := range indexChan {
		index.Index(v.Path, v)
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
