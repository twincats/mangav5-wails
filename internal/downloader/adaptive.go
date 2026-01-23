package downloader

import (
	"sync"
	"time"
)

type AdaptiveController struct {
	current int
	min     int
	max     int

	window []time.Duration
	size   int
	mu     sync.Mutex
}

func NewAdaptiveController(start, min, max int) *AdaptiveController {
	return &AdaptiveController{
		current: start,
		min:     min,
		max:     max,
		size:    5,
	}
}

func (c *AdaptiveController) AddLatency(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.window = append(c.window, d)
	if len(c.window) > c.size {
		c.window = c.window[1:]
	}
}

func (c *AdaptiveController) Adjust(success bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !success {
		c.current = max(c.min, c.current/2)
		return
	}

	if len(c.window) < c.size {
		return
	}

	var sum time.Duration
	for _, d := range c.window {
		sum += d
	}
	avg := sum / time.Duration(len(c.window))

	switch {
	case avg < 400*time.Millisecond && c.current < c.max:
		c.current++
	case avg > 800*time.Millisecond:
		c.current = max(c.min, c.current/2)
	}
}

func (c *AdaptiveController) Current() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.current
}
