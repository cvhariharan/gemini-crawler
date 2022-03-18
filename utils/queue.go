package utils

import (
	"container/list"
	"fmt"
	"sync"
)

type Queue struct {
	Q     *list.List
	M     sync.Mutex
	Added map[string]bool
}

func NewQueue() *Queue {
	return &Queue{
		Q:     list.New(),
		Added: make(map[string]bool),
	}
}

func (q *Queue) Enqueue(path string) {
	q.M.Lock()
	defer q.M.Unlock()
	q.Q.PushBack(path)
	q.Added[path] = true
}

func (q *Queue) Dequeue() string {
	q.M.Lock()
	defer q.M.Unlock()
	e := q.Q.Front()
	val, ok := e.Value.(string)
	q.Q.Remove(e)
	if ok {
		return val
	}
	return ""
}

func (q *Queue) PrintAll() {
	for e := q.Q.Front(); e != nil; e = e.Next() {
		fmt.Println(e.Value)
	}
}
