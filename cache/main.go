package cache

type node struct {
	key, value string
	prev, next *node
}


func newNode(key, value string) *node {
	return &node{key: key, value: value}
}

type dll struct {
	head *node
	tail *node
	m    map[string]*node
}

func newDLL() *dll {
	head := newNode("", "")
	tail := newNode("", "")
	head.next = tail
	tail.prev = head
	return &dll{
		head: head,
		tail: tail,
		m:    make(map[string]*node),
	}
}

func (l *dll) addToFront(key, value string) {
	nextNode := l.head.next
	newNode := newNode(key, value)
	l.head.next = newNode
	newNode.prev = l.head
	newNode.next = nextNode
	nextNode.prev = newNode
	l.m[key] = newNode
}

func (l *dll) deleteLastNode() {
	if l.tail.prev == l.head {
		return
	}
	delete(l.m, l.tail.prev.key)
	prevToPrev := l.tail.prev.prev
	prevToPrev.next = l.tail
	l.tail.prev = prevToPrev
}

func (l *dll) deleteNode(n *node) {
	delete(l.m, n.key)
	left := n.prev
	right := n.next
	left.next = right
	right.prev = left
}

func (l *dll) get(key string) (string, bool) {
	if n, ok := l.m[key]; ok {
		val := n.value
		l.deleteNode(n)
		l.addToFront(key, val)
		return val, true
	}
	return "", false
}

func (l *dll) addNode(key, value string, capacity int) {
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

type LRUCache struct {
	capacity int
	list     *dll
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		list:     newDLL(),
	}
}

func (c *LRUCache) Get(key string) (string, bool) {
	return c.list.get(key)
}

func (c *LRUCache) Put(key, value string) {
	c.list.addNode(key, value, c.capacity)
}
func (c *LRUCache) DeleteKey(key string) {
	if n, ok := c.list.m[key]; ok {
		c.list.deleteNode(n)
	}
}