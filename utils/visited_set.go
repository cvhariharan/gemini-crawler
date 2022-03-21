package utils

import (
	"log"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
)

type Data struct {
	Path string
	Text string
}

type VisitedSet interface {
	IsVisited(string) bool
	Visit(string) error
	IsIndexed(string) bool
	Index(string) error
	GetNotIndexed(chan string) error
	Close()
}

type PersistentSet struct {
	db      *leveldb.DB
	logger  *log.Logger
	logFile *os.File
}

func NewIndexSet() VisitedSet {
	db, err := leveldb.OpenFile("kvdb", nil)
	if err != nil {
		log.Fatal(err)
	}

	logFile, err := os.Create("indexed.txt")
	if err != nil {
		log.Fatal(err)
	}
	logger := log.New(logFile, "", 0)

	return &PersistentSet{
		db:      db,
		logger:  logger,
		logFile: logFile,
	}
}

// IsIndexer returns true if the contents of the page are added to the index (bleve)
func (p *PersistentSet) IsIndexed(path string) bool {
	data, err := p.db.Get([]byte(path), nil)
	if err != nil && err != leveldb.ErrNotFound {
		log.Println(err)
	}
	return string(data) == "2"
}

func (p *PersistentSet) Index(path string) error {
	p.logger.Println(path)
	return p.db.Put([]byte(path), []byte("2"), nil)
}

func (p *PersistentSet) IsVisited(path string) bool {
	data, err := p.db.Get([]byte(path), nil)
	if err != nil && err != leveldb.ErrNotFound {
		log.Println(err)
	}
	return string(data) == "1"
}

func (p *PersistentSet) Visit(path string) error {
	return p.db.Put([]byte(path), []byte("1"), nil)
}

func (p *PersistentSet) GetNotIndexed(indexChan chan string) error {
	iter := p.db.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		if string(value) == "1" {
			log.Println(string(key))
			indexChan <- string(key)
		}
	}
	iter.Release()
	return iter.Error()
}

func (p *PersistentSet) Close() {
	p.db.Close()
	p.logFile.Close()
}
