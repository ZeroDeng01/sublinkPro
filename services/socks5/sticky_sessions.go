package socks5

import (
	"sync"
	"time"

	"sublink/models"
)

const maxStickySessionLeases = 10000

type stickySessionLease struct {
	nodeKey   string
	expiresAt time.Time
}

type stickySessionTable struct {
	mu     sync.Mutex
	leases map[string]stickySessionLease
	now    func() time.Time
}

func newStickySessionTable() *stickySessionTable {
	return &stickySessionTable{leases: make(map[string]stickySessionLease), now: time.Now}
}

func (t *stickySessionTable) promote(key string, nodes []models.Node) ([]models.Node, bool) {
	if key == "" || len(nodes) == 0 {
		return nodes, false
	}
	now := t.now()
	t.mu.Lock()
	defer t.mu.Unlock()
	lease, ok := t.leases[key]
	if ok && !lease.expiresAt.After(now) {
		delete(t.leases, key)
		ok = false
	}
	if !ok {
		return nodes, false
	}

	for index, node := range nodes {
		if adapterKey(node) != lease.nodeKey {
			continue
		}
		if index == 0 {
			return nodes, true
		}
		promoted := make([]models.Node, 0, len(nodes))
		promoted = append(promoted, node)
		promoted = append(promoted, nodes[:index]...)
		promoted = append(promoted, nodes[index+1:]...)
		return promoted, true
	}

	delete(t.leases, key)
	return nodes, false
}

func (t *stickySessionTable) bind(key string, node models.Node, ttl time.Duration) {
	if key == "" || ttl <= 0 {
		return
	}
	now := t.now()
	t.mu.Lock()
	_, refreshing := t.leases[key]
	if !refreshing && len(t.leases) >= maxStickySessionLeases {
		t.pruneLocked(now)
	}
	if !refreshing && len(t.leases) >= maxStickySessionLeases {
		var oldestKey string
		var oldestExpiry time.Time
		for candidateKey, lease := range t.leases {
			if oldestExpiry.IsZero() || lease.expiresAt.Before(oldestExpiry) {
				oldestKey = candidateKey
				oldestExpiry = lease.expiresAt
			}
		}
		delete(t.leases, oldestKey)
	}
	t.leases[key] = stickySessionLease{nodeKey: adapterKey(node), expiresAt: now.Add(ttl)}
	t.mu.Unlock()
}

func (t *stickySessionTable) remove(key string) {
	if key == "" {
		return
	}
	t.mu.Lock()
	delete(t.leases, key)
	t.mu.Unlock()
}

func (t *stickySessionTable) pruneLocked(now time.Time) {
	for key, lease := range t.leases {
		if !lease.expiresAt.After(now) {
			delete(t.leases, key)
		}
	}
}

func stickySessionKey(cfg Config, clientAddress, username string) string {
	if !cfg.StickySessionEnabled || cfg.Selection == "specific" {
		return ""
	}
	switch cfg.StickySessionMode {
	case "username":
		if username == "" {
			return ""
		}
		return "username:" + username
	default:
		if clientAddress == "" {
			return ""
		}
		return "client_ip:" + clientAddress
	}
}
