package utils

import (
	"strings"

	"github.com/peterbourgon/diskv/v3"
)

type VisitedSet interface {
	IsVisited(string) bool
	Visit(string) error
}

type PersistentSet struct {
	kv *diskv.Diskv
}

func NewIndexSet() VisitedSet {
	d := diskv.New(diskv.Options{
		BasePath:          "kvstore",
		AdvancedTransform: AdvancedTransform,
		InverseTransform:  InverseTransform,
		CacheSizeMax:      1024 * 1024,
	})

	return &PersistentSet{
		kv: d,
	}
}

func (p *PersistentSet) IsVisited(path string) bool {
	return p.kv.Has(path)
}

func (p *PersistentSet) Visit(path string) error {
	return p.kv.Write(path, []byte("1"))
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