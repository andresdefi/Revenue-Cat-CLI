package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	cacheDir   = ".rc"
	cacheSub   = "cache"
	defaultTTL = 5 * time.Minute
)

type entry struct {
	Data      json.RawMessage `json:"data"`
	ExpiresAt int64           `json:"expires_at"`
}

// Get retrieves a cached value. Returns nil if not found or expired.
func Get(key string) []byte {
	e, err := load(key)
	if err != nil {
		return nil
	}
	if time.Now().Unix() > e.ExpiresAt {
		_ = os.Remove(path(key))
		return nil
	}
	return e.Data
}

// Set stores a value in the cache with the default TTL.
func Set(key string, data []byte) {
	SetWithTTL(key, data, defaultTTL)
}

// SetWithTTL stores a value in the cache with a custom TTL.
func SetWithTTL(key string, data []byte, ttl time.Duration) {
	dir := dir()
	if dir == "" {
		return
	}
	_ = os.MkdirAll(dir, 0o700)

	e := entry{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl).Unix(),
	}
	raw, err := json.Marshal(e)
	if err != nil {
		return
	}
	_ = os.WriteFile(path(key), raw, 0o600)
}

// Clear removes all cached entries.
func Clear() error {
	d := dir()
	if d == "" {
		return nil
	}
	return os.RemoveAll(d)
}

func load(key string) (*entry, error) {
	data, err := os.ReadFile(path(key))
	if err != nil {
		return nil, err
	}
	var e entry
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

func path(key string) string {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
	return filepath.Join(dir(), hash[:16]+".json")
}

func dir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, cacheDir, cacheSub)
}
