package cache

import "sync"

type node struct {
	key, value string
	prev, next *node
}

type dll struct {
	head *node
	tail *node
	m    map[string]*node
}

func newDLL() *dll {
	head := &node{}
	tail := &node{}
	head.next = tail
	tail.prev = head
	return &dll{head: head, tail: tail, m: make(map[string]*node)}
}

func (l *dll) moveToFront(n *node) {
	// detach
	n.prev.next = n.next
	n.next.prev = n.prev
	// insert after head
	n.next = l.head.next
	n.prev = l.head
	l.head.next.prev = n
	l.head.next = n
}

func (l *dll) addToFront(key, value string) *node {
	n := &node{key: key, value: value}
	n.next = l.head.next
	n.prev = l.head
	l.head.next.prev = n
	l.head.next = n
	l.m[key] = n
	return n
}

func (l *dll) removeLast() {
	last := l.tail.prev
	if last == l.head {
		return
	}
	delete(l.m, last.key)
	last.prev.next = l.tail
	l.tail.prev = last.prev
}

type LRUCache struct {
	mu       sync.RWMutex
	capacity int
	list     *dll
}

func NewLRUCache(cap int) *LRUCache {
	return &LRUCache{capacity: cap, list: newDLL()}
}

// Highly Optimized GET
func (c *LRUCache) Get(key string) (string, bool) {
	// First do fast path under read lock
	c.mu.RLock()
	n, ok := c.list.m[key]
	c.mu.RUnlock()
	if !ok {
		return "", false // miss → no write-lock needed
	}

	val := n.value

	/*
		Key already exists → must move it to front
		Requires write lock
	*/
	c.mu.Lock()
	// Re-check to avoid race between locks
	if n2, exists := c.list.m[key]; exists && n2 == n {
		c.list.moveToFront(n)
	}
	c.mu.Unlock()

	return val, true
}

func (c *LRUCache) Put(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if n, ok := c.list.m[key]; ok {
		n.value = value
		c.list.moveToFront(n)
		return
	}

	if len(c.list.m) >= c.capacity {
		c.list.removeLast()
	}

	c.list.addToFront(key, value)
}

func (c *LRUCache) DeleteKey(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if n, ok := c.list.m[key]; ok {
		n.prev.next = n.next
		n.next.prev = n.prev
		delete(c.list.m, key)
	}
}
