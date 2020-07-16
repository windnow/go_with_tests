package main

import "sync"

// Counter ...
type Counter struct {
	mu    sync.Mutex
	value int
}

// Inc ...
func (c *Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

// Value ...
func (c *Counter) Value() int {
	return c.value
}

// NewCounter create pointer to new counter
func NewCounter() *Counter {
	return &Counter{}
}
