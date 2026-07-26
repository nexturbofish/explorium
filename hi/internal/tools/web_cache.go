package tools

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type WebCache struct {
	mu         sync.RWMutex
	maxEntries int
	items      map[string]webCacheItem
}

func NewWebCache() *WebCache {
	return &WebCache{items: make(map[string]webCacheItem), maxEntries: 100}
}

type webCacheItem struct {
	result    string
	expiresAt time.Time
}

func (c *WebCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.items[key]
	if !ok || time.Now().After(item.expiresAt) {
		return "", false
	}
	return item.result, true
}

func (c *WebCache) Set(key, result string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.items) >= c.maxEntries {
		// 简单淘汰策略：删除最早国旗的条目
		var oldestKey string
		var oldest time.Time
		for k, v := range c.items {
			if oldestKey == "" || v.expiresAt.Before(oldest) {
				oldestKey = k
				oldest = v.expiresAt
			}
		}
		delete(c.items, oldestKey)
	}
	c.items[key] = webCacheItem{result: result, expiresAt: time.Now().Add(ttl)}
}

type WebFetcher interface {
	Fetch(ctx context.Context, url string) (string, error)
}

type WebSearcher interface {
	Search(ctx context.Context, query string) (string, error)
}

type WebToolsContext struct {
	cache    *WebCache
	fetcher  WebFetcher
	searcher WebSearcher
	cacheTTL time.Duration
}

func NewWebToolsContext() *WebToolsContext {
	return &WebToolsContext{
		cache:    NewWebCache(),
		cacheTTL: 5 * time.Minute,
	}
}

func (wtc *WebToolsContext) WithFetcher(f WebFetcher) *WebToolsContext {
	wtc.fetcher = f
	return wtc
}

func (wtc *WebToolsContext) WithSearcher(f WebSearcher) *WebToolsContext {
	wtc.searcher = f
	return wtc
}

func (wtc *WebToolsContext) Fetch(ctx context.Context, url string) (string, error) {
	if cached, ok := wtc.cache.Get(url); ok {
		return cached, nil
	}
	if wtc.fetcher == nil {
		return "", fmt.Errorf("web fetching is not configured")
	}
	result, err := wtc.fetcher.Fetch(ctx, url)
	if err != nil {
		return "", err
	}
	wtc.cache.Set(url, result, wtc.cacheTTL)
	return result, nil
}

func (wtc *WebToolsContext) Search(ctx context.Context, query string) (string, error) {
	if cached, ok := wtc.cache.Get(query); ok {
		return cached, nil
	}
	if wtc.searcher == nil {
		return "", fmt.Errorf("web search is not configured")
	}
	result, err := wtc.searcher.Search(ctx, query)
	if err != nil {
		return "", err
	}
	wtc.cache.Set(query, result, wtc.cacheTTL)
	return result, nil
}
