package socks5

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"sync"

	"sublink/models"
	"sublink/services/mihomo"

	"github.com/metacubex/mihomo/constant"
)

const defaultAdapterPoolSize = 64

type adapterFactory func(node models.Node) (constant.Proxy, error)

type adapterEntry struct {
	adapter  constant.Proxy
	lastUsed uint64
	active   int
	retired  bool
	closed   bool
}

type adapterLease struct {
	pool  *adapterPool
	key   string
	entry *adapterEntry
	once  sync.Once
}

type adapterPool struct {
	mu      sync.Mutex
	entries map[string]*adapterEntry
	factory adapterFactory
	closed  bool
	clock   uint64
	maxSize int
}

func newAdapterPool(factory adapterFactory) *adapterPool {
	if factory == nil {
		factory = func(node models.Node) (constant.Proxy, error) {
			return mihomo.GetMihomoAdapter(node.Link)
		}
	}
	return &adapterPool{entries: make(map[string]*adapterEntry), factory: factory, maxSize: defaultAdapterPoolSize}
}

func adapterKey(node models.Node) string {
	hash := node.LinkHash
	if hash == "" {
		sum := sha256.Sum256([]byte(node.Link))
		hash = hex.EncodeToString(sum[:])
	}
	return fmt.Sprintf("%d:%s", node.ID, hash)
}

func adapterNodePrefix(node models.Node) string {
	return fmt.Sprintf("%d:", node.ID)
}

func (p *adapterPool) acquire(node models.Node) (constant.Proxy, *adapterLease, error) {
	key := adapterKey(node)
	prefix := adapterNodePrefix(node)

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, nil, fmt.Errorf("SOCKS5 adapter pool is closed")
	}
	p.clock++
	if entry := p.entries[key]; entry != nil {
		entry.lastUsed = p.clock
		entry.active++
		p.mu.Unlock()
		return entry.adapter, &adapterLease{pool: p, key: key, entry: entry}, nil
	}

	toClose := make([]constant.Proxy, 0, 2)
	for oldKey, entry := range p.entries {
		if oldKey != key && strings.HasPrefix(oldKey, prefix) {
			delete(p.entries, oldKey)
			if adapter := retireEntryLocked(entry); adapter != nil {
				toClose = append(toClose, adapter)
			}
		}
	}
	adapter, err := p.factory(node)
	if err != nil {
		p.mu.Unlock()
		closeAdapters(toClose)
		return nil, nil, err
	}
	entry := &adapterEntry{adapter: adapter, lastUsed: p.clock, active: 1}
	p.entries[key] = entry
	toClose = append(toClose, p.evictOldestLocked(key)...)
	p.mu.Unlock()
	closeAdapters(toClose)
	return adapter, &adapterLease{pool: p, key: key, entry: entry}, nil
}

func (p *adapterPool) evictOldestLocked(excludeKey string) []constant.Proxy {
	toClose := make([]constant.Proxy, 0, 1)
	for len(p.entries) > p.maxSize {
		oldestKey := ""
		var oldestUse uint64
		for key, entry := range p.entries {
			if key == excludeKey || entry.active > 0 {
				continue
			}
			if oldestKey == "" || entry.lastUsed < oldestUse {
				oldestKey, oldestUse = key, entry.lastUsed
			}
		}
		if oldestKey == "" {
			break
		}
		entry := p.entries[oldestKey]
		delete(p.entries, oldestKey)
		if adapter := retireEntryLocked(entry); adapter != nil {
			toClose = append(toClose, adapter)
		}
	}
	return toClose
}

func (lease *adapterLease) release(failed bool) {
	if lease == nil {
		return
	}
	lease.once.Do(func() {
		p := lease.pool
		p.mu.Lock()
		entry := lease.entry
		if failed {
			if current := p.entries[lease.key]; current == entry {
				delete(p.entries, lease.key)
			}
			entry.retired = true
		}
		if entry.active > 0 {
			entry.active--
		}
		var toClose []constant.Proxy
		if entry.retired && entry.active == 0 && !entry.closed {
			entry.closed = true
			toClose = append(toClose, entry.adapter)
		}
		toClose = append(toClose, p.evictOldestLocked("")...)
		p.mu.Unlock()
		closeAdapters(toClose)
	})
}

func retireEntryLocked(entry *adapterEntry) constant.Proxy {
	entry.retired = true
	if entry.active == 0 && !entry.closed {
		entry.closed = true
		return entry.adapter
	}
	return nil
}

func closeAdapters(adapters []constant.Proxy) {
	for _, adapter := range adapters {
		_ = adapter.Close()
	}
}

func (p *adapterPool) close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	toClose := make([]constant.Proxy, 0, len(p.entries))
	for key, entry := range p.entries {
		delete(p.entries, key)
		if adapter := retireEntryLocked(entry); adapter != nil {
			toClose = append(toClose, adapter)
		}
	}
	p.mu.Unlock()
	closeAdapters(toClose)
}

type pooledConn struct {
	net.Conn
	lease *adapterLease
}

func (c *pooledConn) Close() error {
	err := c.Conn.Close()
	c.lease.release(false)
	return err
}
