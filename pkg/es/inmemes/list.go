package es

import "sync"

type Node struct {
	prev *Node
	next *Node
	Data interface{}
}

type List struct {
	headMx sync.RWMutex
	head   *Node
	tail   *Node
	len    uint64
}

func NewList() *List {
	return &List{}
}

// Add element to head
func (l *List) PushToHead(data interface{}) *Node {
	l.headMx.Lock()
	defer l.headMx.Unlock()
	nn := &Node{
		prev: l.head,
		next: nil,
		Data: data,
	}
	if l.head != nil {
		l.head.next = nn
	}
	l.head = nn
	if l.tail == nil {
		l.tail = l.head
	}
	l.len++
	return nn
}

// Forward iteration with additional synchronization on the head, as head may be changed during iteration
func (l *List) Iterate(forEachFunc func(data interface{}, head bool) bool) {
	cur := l.tail
	for cur != nil {
		next := cur.next
		if next == nil {
			// head
			l.headMx.RLock() // maybe wait head change
			next = cur.next
			if next != nil { // become not a head
				l.headMx.RUnlock() // unlock before callback as cur is not a head
			}
		}
		cont := forEachFunc(cur.Data, next == nil)
		if next == nil {
			l.headMx.RUnlock() // unlock after callback as cur is a head
		}
		if !cont {
			return
		}
		cur = next
	}
}

// Reverse is simple iteration, as no modification can be executed on process, as no push tail method exists
func (l *List) IterateReverse(forEachFunc func(data interface{}) bool) {
	l.headMx.RLock()
	cur := l.head
	l.headMx.RUnlock()
	for cur != nil {
		if cont := forEachFunc(cur.Data); !cont {
			return
		}
		cur = cur.prev
	}
}

func (l *List) GetLastElement() interface{} {
	l.headMx.RLock()
	defer l.headMx.RUnlock()
	if l.head == nil {
		return nil
	}
	return l.head.Data
}

func (l *List) Len() uint64 {
	l.headMx.RLock()
	defer l.headMx.RUnlock()
	return l.len
}
