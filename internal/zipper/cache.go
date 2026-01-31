package zipper

import (
	"sync"

	"github.com/klauspost/compress/zip"
)

// globalZipCache is a singleton instance of our cache
var globalZipCache *ZipCache
var once sync.Once

// GetZipCache returns the singleton instance of ZipCache
func GetZipCache() *ZipCache {
	once.Do(func() {
		// Capacity 10 is enough for most manga reading scenarios (prev, current, next chapters)
		globalZipCache = NewZipCache(10)
	})
	return globalZipCache
}

// ZipCache implements a simple Thread-Safe LRU Cache for open zip readers.
// It keeps the file handles open to avoid parsing the Central Directory repeatedly.
type ZipCache struct {
	capacity int
	cache    map[string]*zip.ReadCloser
	order    []string // Keeps track of access order (Front = Most Recent, Back = Oldest)
	mu       sync.Mutex
}

func NewZipCache(capacity int) *ZipCache {
	return &ZipCache{
		capacity: capacity,
		cache:    make(map[string]*zip.ReadCloser),
		order:    make([]string, 0, capacity),
	}
}

// GetOrOpen returns an existing zip reader or opens a new one.
func (c *ZipCache) GetOrOpen(archivePath string) (*zip.ReadCloser, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. Check if exists in cache
	if reader, ok := c.cache[archivePath]; ok {
		c.moveToFront(archivePath)
		return reader, nil
	}

	// 2. Open new reader if not found
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, err
	}

	// 3. Add to cache
	// If full, evict the oldest
	if len(c.order) >= c.capacity {
		c.evictOldest()
	}

	c.cache[archivePath] = reader
	// Prepend to order (add to front)
	c.order = append([]string{archivePath}, c.order...)

	return reader, nil
}

// Remove removes an item from the cache and closes its reader.
// This is useful when the underlying file needs to be modified or deleted.
func (c *ZipCache) Remove(archivePath string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if exists
	if reader, ok := c.cache[archivePath]; ok {
		reader.Close()
		delete(c.cache, archivePath)
		
		// Remove from order slice
		for i, k := range c.order {
			if k == archivePath {
				c.order = append(c.order[:i], c.order[i+1:]...)
				break
			}
		}
	}
}

// moveToFront moves the key to the beginning of the order slice
func (c *ZipCache) moveToFront(key string) {
	// Find index
	idx := -1
	for i, k := range c.order {
		if k == key {
			idx = i
			break
		}
	}

	if idx == -1 || idx == 0 {
		return // Not found (shouldn't happen) or already at front
	}

	// Remove from current position
	c.order = append(c.order[:idx], c.order[idx+1:]...)
	// Prepend
	c.order = append([]string{key}, c.order...)
}

// evictOldest removes and closes the oldest accessed item
func (c *ZipCache) evictOldest() {
	if len(c.order) == 0 {
		return
	}

	// Oldest is at the end
	lastIdx := len(c.order) - 1
	key := c.order[lastIdx]

	// Close the reader
	if reader, ok := c.cache[key]; ok {
		reader.Close()
	}

	// Remove from map and slice
	delete(c.cache, key)
	c.order = c.order[:lastIdx]
}

// CloseAll closes all open zip readers and clears the cache
func (c *ZipCache) CloseAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, reader := range c.cache {
		reader.Close()
	}
	c.cache = make(map[string]*zip.ReadCloser)
	c.order = c.order[:0]
}
