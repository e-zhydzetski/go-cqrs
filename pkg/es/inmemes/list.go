package inmemes

import (
	"sync"
)

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
// If no elements lost - finalizer func invoked, push blocked during finalizer working
func (l *List) Iterate(forEachFunc func(data interface{}) bool, finalizerFunc func()) {
	cur := l.tail
	for {
		if cur != nil { // has element
			if cont := forEachFunc(cur.Data); !cont {
				return
			}
		}
		if cur == l.head && finalizerFunc != nil { // finalization on head if needed
			cur = func() *Node {
				l.headMx.RLock() // maybe wait head change
				defer l.headMx.RUnlock()
				if cur == l.head { // head not changed
					finalizerFunc()
					return nil // iteration should be terminated after finalization
				}
				// element added
				if cur == nil { // head was nil, list was empty, so start from the beginning
					return &Node{ // fake node before tail, to get tail on next step
						next: l.tail,
					}
				}
				return cur // head was not nil, so simple continue iteration
			}()
		}
		if cur == nil {
			break
		}
		cur = cur.next
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
