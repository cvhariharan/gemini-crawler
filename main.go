package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/cvhariharan/gemini-crawler/gemini"
	"github.com/cvhariharan/gemini-crawler/gemtext"
	"github.com/cvhariharan/gemini-crawler/utils"
)

const (
	WORKING_INDEX = "index2"
	INDEX         = "index"
	SEED_FILE     = "seeds.txt"
	WORKERS       = 5
)

var LinkLock sync.Mutex
var IndexOnly = flag.Bool("index", false, "Set index only, no new URLs will be added to the queue")

func main() {
	flag.Parse()

	urlLog, err := os.Create("urls.txt")
	if err != nil {
		panic(err)
	}
	defer urlLog.Close()
	mw := io.MultiWriter(os.Stdout, urlLog)
	urlLogger := log.New(mw, "", 0)

	var wg sync.WaitGroup
	mountPoint := os.Getenv("AWS_EFS_MOUNT")
	if mountPoint == "" {
		log.Fatal("mount point not set. AWS_EFS_MOUNT empty")
	}

	// Create a new index at mountpoint/index2
	workingIndex := "index-" + time.Now().Format("2006-01-02-150405")
	indexPath := filepath.Join(mountPoint, workingIndex)
	fmt.Println("Index path -", indexPath)

	c := make(chan string, 200)
	q := utils.NewIndexSet()
	defer q.Close()

	removeIfExists(indexPath)

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New(indexPath, mapping)
	if err != nil {
		log.Fatal(err)
	}

	if *IndexOnly {
		fmt.Println("Indexing only")
		var wg1 sync.WaitGroup
		toIndex := make(chan string, 20)
		defer close(toIndex)

		wg1.Add(1)
		go indexFeeder(toIndex, index, q, &wg1)

		q.GetNotIndexed(toIndex)

		wg1.Wait()
		return
	}

	indexChan := make(chan utils.Data, 200)
	defer close(indexChan)

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

	start := time.Now()
	go indexer(indexChan, index, q, &wg)
	for i := 0; i < WORKERS; i++ {
		wg.Add(1)
		go createCrawler(c, indexChan, q, urlLogger, &wg)
	}

	for _, v := range seeds {
		c <- v
	}

	wg.Wait()

	end := time.Now()

	fmt.Printf("Indexing complete in %f minutes\n", end.Sub(start).Minutes())
}

func createCrawler(c chan string, indexChan chan utils.Data, q utils.VisitedSet, urlLogger *log.Logger, wg *sync.WaitGroup) {
	client := gemini.NewClient(gemini.ClientOptions{Insecure: true})
	for path := range c {
		if q.IsIndexed(path) {
			continue
		}
		urlLogger.Println(path)
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
					if !q.IsIndexed(v) {
						q.Visit(v)
						c <- v
					}
				}(v)
			}

			go func() {
				indexChan <- utils.Data{Path: path, Text: string(txt)}
			}()
		}

		resp.Close()
	}
	wg.Done()
}

func indexFeeder(toIndex chan string, index bleve.Index, q utils.VisitedSet, wg *sync.WaitGroup) {
	indexChan := make(chan utils.Data)
	defer close(indexChan)

	client := gemini.NewClient(gemini.ClientOptions{Insecure: true})

	go indexer(indexChan, index, q, wg)
	for url := range toIndex {
		if q.IsIndexed(url) {
			continue
		}
		resp, err := client.Fetch(url)
		if err != nil {
			log.Println(err)
			continue
		}
		if resp.Meta == "text/gemini" {
			txt, _ := ioutil.ReadAll(resp.Body)
			indexChan <- utils.Data{Path: url, Text: string(txt)}
		}
		resp.Close()
	}

	wg.Done()
}

func indexer(indexChan chan utils.Data, index bleve.Index, q utils.VisitedSet, wg *sync.WaitGroup) {
	for v := range indexChan {
		index.Index(v.Path, v)
		q.Index(v.Path)
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
