package es

type Node struct {
	prev *Node
	next *Node
	Data interface{}
}

type List struct {
	head *Node
	tail *Node
}

func NewList() *List {
	return &List{}
}

// Add element to head
func (l *List) PushToHead(data interface{}) *Node {
	nn := &Node{
		prev: l.head,
		next: nil,
		Data: data,
	}
	l.head = nn
	if l.tail == nil {
		l.tail = l.head
	}
	return nn
}

// Invoke callback func for each element, stop on callback return false
func (l *List) Iterate(forEachFunc func(data interface{}) bool, reverse bool) {
	cur := l.tail
	if reverse {
		cur = l.head
	}
	for cur != nil {
		if cont := forEachFunc(cur.Data); !cont {
			return
		}
		if reverse {
			cur = cur.prev
		} else {
			cur = cur.next
		}
	}
}

func (l *List) GetLastElement() interface{} {
	if l.head == nil {
		return nil
	}
	return l.head.Data
}
