package main

import "fmt"

// Node is unexported, so only code in this package can access its fields
type node struct {
	key, value int
	prev, next *node
}

// newNode acts as a "friend" function — it can access private fields
func newNode(key, value int) *node {
	return &node{key: key, value: value}
}

// dll (doubly linked list + map)
type dll struct {
	head *node
	tail *node
	m    map[int]*node
}

// newDLL is another friend function, initializing dummy head/tail
func newDLL() *dll {
	head := newNode(-1, -1)
	tail := newNode(-1, -1)
	head.next = tail
	tail.prev = head
	return &dll{
		head: head,
		tail: tail,
		m:    make(map[int]*node),
	}
}

// addToFront adds a node right after head
func (l *dll) addToFront(key, value int) {
	nextNode := l.head.next
	newNode := newNode(key, value)
	l.head.next = newNode
	nextNode.prev = newNode
	newNode.prev = l.head
	newNode.next = nextNode
	l.m[key] = newNode
}

// deleteLastNode removes the least-recently-used node (before tail)
func (l *dll) deleteLastNode() {
	if l.tail.prev == l.head {
		return // nothing to delete
	}
	delete(l.m, l.tail.prev.key)
	prevToPrev := l.tail.prev.prev
	prevToPrev.next = l.tail
	l.tail.prev = prevToPrev
}

// deleteNode removes a specific node from the list
func (l *dll) deleteNode(n *node) {
	delete(l.m, n.key)
	left := n.prev
	right := n.next
	left.next = right
	right.prev = left
}

// get retrieves a value by key and moves the node to front
func (l *dll) get(key int) int {
	if n, ok := l.m[key]; ok {
		val := n.value
		l.deleteNode(n)
		l.addToFront(key, val)
		return val
	}
	return -1
}

// addNode adds or updates a key; if full, removes LRU
func (l *dll) addNode(key, value, capacity int) {
	if n, ok := l.m[key]; ok {
		l.deleteNode(n)
		l.addToFront(key, value)
	} else if len(l.m) < capacity {
		l.addToFront(key, value)
	} else {
		l.deleteLastNode()
		l.addToFront(key, value)
	}
}

// LRUCache is the public interface — this struct can be exported
type LRUCache struct {
	capacity int
	list     *dll
}

// NewLRUCache is the exported constructor
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		list:     newDLL(),
	}
}

func (c *LRUCache) Get(key int) int {
	return c.list.get(key)
}

func (c *LRUCache) Put(key, value int) {
	c.list.addNode(key, value, c.capacity)
}