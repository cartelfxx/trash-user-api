package cache

import (
	"log"
	"sync"
	"time"

	"discord-user-api/models"
	"discord-user-api/websocket"
)

type CacheEntry struct {
	Data      interface{}
	Timestamp time.Time
	TTL       time.Duration
	Hits      int
	LastRefresh time.Time
	AutoRefresh bool
	RefreshInterval time.Duration
}

type Cache struct {
	data            map[string]*CacheEntry
	mutex           sync.RWMutex
	maxSize         int
	defaultTTL      time.Duration
	cleanupInterval time.Duration
	stats           *CacheStats
	stopCleanup     chan bool
	wsManager       *websocket.WebSocketManager
	refreshTicker   *time.Ticker
	stopRefresh     chan bool
}

type CacheStats struct {
	Hits        int64
	Misses      int64
	Evictions   int64
	Size        int
	LastCleanup time.Time
	Refreshes   int64
}

func NewCache(maxSize int, defaultTTL, cleanupInterval time.Duration) *Cache {
	cache := &Cache{
		data:            make(map[string]*CacheEntry),
		maxSize:         maxSize,
		defaultTTL:      defaultTTL,
		cleanupInterval: cleanupInterval,
		stats:           &CacheStats{},
		stopCleanup:     make(chan bool),
		stopRefresh:     make(chan bool),
	}

	go cache.cleanupRoutine()

	log.Printf("💾 Cache başlatıldı (Max: %d, TTL: %v, Cleanup: %v)", maxSize, defaultTTL, cleanupInterval)
	return cache
}

func (c *Cache) SetWebSocketManager(wsManager *websocket.WebSocketManager) {
	c.wsManager = wsManager
	log.Printf("🔌 WebSocket Manager cache'e bağlandı")
}

func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.SetWithAutoRefresh(key, value, ttl, false, 0)
}

func (c *Cache) SetWithAutoRefresh(key string, value interface{}, ttl time.Duration, autoRefresh bool, refreshInterval time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(c.data) >= c.maxSize && c.data[key] == nil {
		c.evictOldest()
	}

	now := time.Now()
	c.data[key] = &CacheEntry{
		Data:           value,
		Timestamp:      now,
		TTL:            ttl,
		Hits:           0,
		LastRefresh:    now,
		AutoRefresh:    autoRefresh,
		RefreshInterval: refreshInterval,
	}

	c.stats.Size = len(c.data)
	
	if c.wsManager != nil {
		c.wsManager.Broadcast(models.WebSocketEvent{
			Type:      "cache_update",
			Data:      models.CacheUpdateEvent{
				Type:      "set",
				Key:       key,
				Timestamp: now.Format(time.RFC3339),
				Data:      value,
			},
			Timestamp: now.Format(time.RFC3339),
		})
	}

	log.Printf("📥 Cache'e eklendi: %s (TTL: %v, AutoRefresh: %t)", key, ttl, autoRefresh)
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	entry, exists := c.data[key]
	c.mutex.RUnlock()

	if !exists {
		c.stats.Misses++
		log.Printf("❌ Cache miss: %s", key)
		return nil, false
	}

	if time.Since(entry.Timestamp) > entry.TTL {
		c.mutex.Lock()
		delete(c.data, key)
		c.stats.Size = len(c.data)
		c.mutex.Unlock()
		c.stats.Misses++
		log.Printf("⏰ Cache expired: %s", key)
		return nil, false
	}

	c.mutex.Lock()
	entry.Hits++
	c.mutex.Unlock()
	c.stats.Hits++

	log.Printf("📤 Cache hit: %s (Hits: %d)", key, entry.Hits)
	return entry.Data, true
}

func (c *Cache) Refresh(key string) {
	c.mutex.RLock()
	entry, exists := c.data[key]
	c.mutex.RUnlock()

	if !exists {
		log.Printf("⚠️ Refresh için öğe bulunamadı: %s", key)
		return
	}

	if !entry.AutoRefresh {
		log.Printf("⚠️ Öğe auto-refresh için yapılandırılmamış: %s", key)
		return
	}

	entry.LastRefresh = time.Now()
	c.stats.Refreshes++

	if c.wsManager != nil {
		c.wsManager.Broadcast(models.WebSocketEvent{
			Type:      "cache_refresh",
			Data:      models.CacheUpdateEvent{
				Type:      "refresh",
				Key:       key,
				Timestamp: time.Now().Format(time.RFC3339),
				Data:      entry.Data,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}

	log.Printf("🔄 Cache öğesi yenilendi: %s", key)
}

func (c *Cache) StartAutoRefresh() {
	if c.refreshTicker != nil {
		return
	}

	c.refreshTicker = time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case <-c.refreshTicker.C:
				c.checkAutoRefresh()
			case <-c.stopRefresh:
				return
			}
		}
	}()

	log.Printf("🔄 Auto-refresh sistemi başlatıldı")
}

func (c *Cache) StopAutoRefresh() {
	if c.refreshTicker != nil {
		c.refreshTicker.Stop()
		c.stopRefresh <- true
		c.refreshTicker = nil
		log.Printf("🛑 Auto-refresh sistemi durduruldu")
	}
}

func (c *Cache) checkAutoRefresh() {
	c.mutex.RLock()
	var toRefresh []string
	now := time.Now()

	for key, entry := range c.data {
		if entry.AutoRefresh && entry.RefreshInterval > 0 {
			if now.Sub(entry.LastRefresh) >= entry.RefreshInterval {
				toRefresh = append(toRefresh, key)
			}
		}
	}
	c.mutex.RUnlock()

	for _, key := range toRefresh {
		c.Refresh(key)
	}

	if len(toRefresh) > 0 {
		log.Printf("🔄 %d öğe otomatik yenilendi", len(toRefresh))
	}
}

func (c *Cache) Delete(key string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.data[key]; exists {
		delete(c.data, key)
		c.stats.Size = len(c.data)
		
		if c.wsManager != nil {
			c.wsManager.Broadcast(models.WebSocketEvent{
				Type:      "cache_delete",
				Data:      models.CacheUpdateEvent{
					Type:      "delete",
					Key:       key,
					Timestamp: time.Now().Format(time.RFC3339),
				},
				Timestamp: time.Now().Format(time.RFC3339),
			})
		}

		log.Printf("🗑️  Cache'den silindi: %s", key)
		return true
	}
	return false
}

func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	count := len(c.data)
	c.data = make(map[string]*CacheEntry)
	c.stats.Size = 0
	
	if c.wsManager != nil {
		c.wsManager.Broadcast(models.WebSocketEvent{
			Type:      "cache_clear",
			Data:      models.CacheUpdateEvent{
				Type:      "clear",
				Timestamp: time.Now().Format(time.RFC3339),
			},
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}

	log.Printf("🧹 Cache temizlendi: %d öğe silindi", count)
}

func (c *Cache) GetStats() *CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	stats := *c.stats
	stats.Size = len(c.data)
	return &stats
}

func (c *Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.data {
		if oldestKey == "" || entry.Timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Timestamp
		}
	}

	if oldestKey != "" {
		delete(c.data, oldestKey)
		c.stats.Evictions++
		log.Printf("🚫 En eski öğe cache'den çıkarıldı: %s", oldestKey)
	}
}

func (c *Cache) cleanupRoutine() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCleanup:
			log.Printf("🛑 Cache cleanup durduruldu")
			return
		}
	}
}

func (c *Cache) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	removed := 0

	for key, entry := range c.data {
		if now.Sub(entry.Timestamp) > entry.TTL {
			delete(c.data, key)
			removed++
		}
	}

	if removed > 0 {
		c.stats.Size = len(c.data)
		c.stats.LastCleanup = now
		log.Printf("🧹 Cache cleanup: %d süresi dolmuş öğe silindi", removed)
	}
}

func (c *Cache) Stop() {
	close(c.stopCleanup)
	c.StopAutoRefresh()
}

func (c *Cache) PrintStats() {
	stats := c.GetStats()
	hitRate := float64(0)
	if stats.Hits+stats.Misses > 0 {
		hitRate = float64(stats.Hits) / float64(stats.Hits+stats.Misses) * 100
	}

	log.Printf("📊 Cache İstatistikleri:")
	log.Printf("   📈 Hit Rate: %.2f%% (%d/%d)", hitRate, stats.Hits, stats.Hits+stats.Misses)
	log.Printf("   📦 Size: %d/%d", stats.Size, c.maxSize)
	log.Printf("   🚫 Evictions: %d", stats.Evictions)
	log.Printf("   🔄 Refreshes: %d", stats.Refreshes)
	log.Printf("   🧹 Last Cleanup: %v", stats.LastCleanup.Format("15:04:05"))
} 