package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

// ContentCache provides intelligent caching for rendered TUI content
type ContentCache struct {
	entries map[string]*CacheEntry
	mutex   sync.RWMutex
	maxSize int
	ttl     time.Duration
}

// CacheEntry represents a cached content entry
type CacheEntry struct {
	Content     string
	Hash        string
	CreatedAt   time.Time
	LastUsed    time.Time
	AccessCount int64
	DataHash    string // Hash of the input data
}

// NewContentCache creates a new content cache
func NewContentCache(maxSize int, ttl time.Duration) *ContentCache {
	cache := &ContentCache{
		entries: make(map[string]*CacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}
	
	// Start cleanup goroutine
	go cache.startCleanupProcess()
	
	return cache
}

// GetOrRender returns cached content or renders new content if cache miss
func (c *ContentCache) GetOrRender(key string, data interface{}, renderer func() string) string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// Generate data hash
	dataHash := c.generateDataHash(data)
	
	// Check for cache hit
	if entry, exists := c.entries[key]; exists {
		// Check if data has changed
		if entry.DataHash == dataHash && time.Since(entry.CreatedAt) < c.ttl {
			// Cache hit - update access info
			entry.LastUsed = time.Now()
			entry.AccessCount++
			return entry.Content
		}
		
		// Data changed or expired - remove old entry
		delete(c.entries, key)
	}
	
	// Cache miss - render new content
	content := renderer()
	contentHash := c.generateContentHash(content)
	
	// Store in cache
	c.entries[key] = &CacheEntry{
		Content:     content,
		Hash:        contentHash,
		CreatedAt:   time.Now(),
		LastUsed:    time.Now(),
		AccessCount: 1,
		DataHash:    dataHash,
	}
	
	// Evict entries if cache is full
	c.evictIfNeeded()
	
	return content
}

// InvalidateByPrefix invalidates all cache entries with keys starting with prefix
func (c *ContentCache) InvalidateByPrefix(prefix string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	keysToDelete := make([]string, 0)
	for key := range c.entries {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			keysToDelete = append(keysToDelete, key)
		}
	}
	
	for _, key := range keysToDelete {
		delete(c.entries, key)
	}
}

// Invalidate removes a specific cache entry
func (c *ContentCache) Invalidate(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.entries, key)
}

// Clear removes all cache entries
func (c *ContentCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.entries = make(map[string]*CacheEntry)
}

// GetStats returns cache statistics
func (c *ContentCache) GetStats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	stats := CacheStats{
		Entries:   len(c.entries),
		MaxSize:   c.maxSize,
		TTL:       c.ttl,
		HitRate:   0.0,
		AvgAge:    0,
		TotalSize: 0,
	}
	
	if len(c.entries) == 0 {
		return stats
	}
	
	totalAccess := int64(0)
	totalAge := time.Duration(0)
	now := time.Now()
	
	for _, entry := range c.entries {
		totalAccess += entry.AccessCount
		totalAge += now.Sub(entry.CreatedAt)
		stats.TotalSize += len(entry.Content)
	}
	
	// Calculate hit rate (entries with > 1 access)
	hits := int64(0)
	for _, entry := range c.entries {
		if entry.AccessCount > 1 {
			hits++
		}
	}
	
	if totalAccess > 0 {
		stats.HitRate = float64(hits) / float64(len(c.entries)) * 100
	}
	
	if len(c.entries) > 0 {
		stats.AvgAge = totalAge / time.Duration(len(c.entries))
	}
	
	return stats
}

// generateDataHash creates a hash from input data
func (c *ContentCache) generateDataHash(data interface{}) string {
	// Convert data to JSON for consistent hashing
	jsonData, err := json.Marshal(data)
	if err != nil {
		// Fallback to string representation
		jsonData = []byte(debugString(data))
	}
	
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])[:16] // Use first 16 chars for efficiency
}

// generateContentHash creates a hash from rendered content
func (c *ContentCache) generateContentHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])[:16]
}

// evictIfNeeded removes old entries if cache is full
func (c *ContentCache) evictIfNeeded() {
	if len(c.entries) <= c.maxSize {
		return
	}
	
	// Find least recently used entries to evict
	type entryInfo struct {
		key      string
		lastUsed time.Time
	}
	
	entries := make([]entryInfo, 0, len(c.entries))
	for key, entry := range c.entries {
		entries = append(entries, entryInfo{
			key:      key,
			lastUsed: entry.LastUsed,
		})
	}
	
	// Sort by last used time (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].lastUsed.After(entries[j].lastUsed) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	
	// Remove oldest entries until we're under max size
	entriesToRemove := len(c.entries) - c.maxSize + 1 // Remove extra to avoid frequent evictions
	for i := 0; i < entriesToRemove && i < len(entries); i++ {
		delete(c.entries, entries[i].key)
	}
}

// startCleanupProcess runs periodic cleanup of expired entries
func (c *ContentCache) startCleanupProcess() {
	ticker := time.NewTicker(c.ttl / 2) // Clean up twice per TTL period
	defer ticker.Stop()
	
	for range ticker.C {
		c.cleanupExpiredEntries()
	}
}

// cleanupExpiredEntries removes expired cache entries
func (c *ContentCache) cleanupExpiredEntries() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	now := time.Now()
	keysToDelete := make([]string, 0)
	
	for key, entry := range c.entries {
		if now.Sub(entry.CreatedAt) > c.ttl {
			keysToDelete = append(keysToDelete, key)
		}
	}
	
	for _, key := range keysToDelete {
		delete(c.entries, key)
	}
}

// CacheStats provides cache performance statistics
type CacheStats struct {
	Entries   int           `json:"entries"`
	MaxSize   int           `json:"max_size"`
	TTL       time.Duration `json:"ttl"`
	HitRate   float64       `json:"hit_rate"`
	AvgAge    time.Duration `json:"avg_age"`
	TotalSize int           `json:"total_size"`
}

// String returns a formatted string representation of cache stats
func (s CacheStats) String() string {
	return debugString(s)
}

// debugString provides a simple string representation of any value
func debugString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	case nil:
		return "<nil>"
	default:
		// Try JSON first
		if jsonBytes, err := json.Marshal(val); err == nil {
			return string(jsonBytes)
		}
		// Fallback to Go's default string representation
		return "<complex-data>"
	}
}