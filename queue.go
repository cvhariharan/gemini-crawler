package main

import (
	"container/list"
	"fmt"
)

type Queue struct {
	Q *list.List
}

func (q *Queue) Enqueue(path string) {
	q.Q.PushBack(path)
}

func (q *Queue) Dequeue() string {
	val, ok := q.Q.Front().Value.(string)
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
