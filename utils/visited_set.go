package utils

import (
	"log"
	"strings"

	"github.com/peterbourgon/diskv/v3"
	"github.com/syndtr/goleveldb/leveldb"
)

type VisitedSet interface {
	IsVisited(string) bool
	Visit(string) error
	IsIndexed(string) bool
	Index(string) error
}

type PersistentSet struct {
	db *leveldb.DB
}

func NewIndexSet() VisitedSet {
	db, err := leveldb.OpenFile("kvdb", nil)
	if err != nil {
		log.Fatal(err)
	}

	return &PersistentSet{
		db: db,
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

func AdvancedTransform(key string) *diskv.PathKey {
	path := strings.Split(key, "/")
	last := len(path) - 1
	return &diskv.PathKey{
		Path:     path[:last],
		FileName: path[last],
	}
}
func InverseTransform(pathKey *diskv.PathKey) (key string) {
	return strings.Join(pathKey.Path, "/") + pathKey.FileName
}
