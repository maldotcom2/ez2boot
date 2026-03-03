package user

import (
	"fmt"
	"sync"
	"time"
)

// In memory cache which keeps track of used MFA codes to prevent re-use

type MFACache struct {
	mu    sync.Mutex
	codes map[string]usedCode
}

type usedCode struct {
	expiry time.Time
}

func NewMFACache() *MFACache {
	c := &MFACache{
		codes: make(map[string]usedCode),
	}
	go c.cleanup()
	return c
}

func (c *MFACache) Has(userID int64, code string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := fmt.Sprintf("%d:%s", userID, code)
	entry, ok := c.codes[key]
	return ok && time.Now().Before(entry.expiry)
}

func (c *MFACache) Set(userID int64, code string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := fmt.Sprintf("%d:%s", userID, code)
	c.codes[key] = usedCode{expiry: time.Now().Add(30 * time.Second)}
}

func (c *MFACache) cleanup() {
	for {
		time.Sleep(time.Minute)
		c.mu.Lock()
		for k, v := range c.codes {
			if time.Now().After(v.expiry) {
				delete(c.codes, k)
			}
		}
		c.mu.Unlock()
	}
}
