package socks5

import (
	"sync"
	"sync/atomic"
	"testing"

	"sublink/models"

	"github.com/metacubex/mihomo/constant"
)

type fakeProxy struct {
	constant.Proxy
	closed atomic.Int32
}

func (f *fakeProxy) Close() error {
	f.closed.Add(1)
	return nil
}

func TestAdapterPoolReusesAndRetiresChangedNodeAfterActiveLeases(t *testing.T) {
	created := 0
	var proxies []*fakeProxy
	pool := newAdapterPool(func(models.Node) (constant.Proxy, error) {
		created++
		proxy := &fakeProxy{}
		proxies = append(proxies, proxy)
		return proxy, nil
	})
	first := models.Node{ID: 7, Link: "socks5://first", LinkHash: "hash-one"}
	a, leaseA, err := pool.acquire(first)
	if err != nil {
		t.Fatal(err)
	}
	b, leaseB, err := pool.acquire(first)
	if err != nil {
		t.Fatal(err)
	}
	if a != b || created != 1 {
		t.Fatalf("adapter was not reused: created=%d", created)
	}
	changed := first
	changed.Link = "socks5://changed"
	changed.LinkHash = "hash-two"
	_, changedLease, err := pool.acquire(changed)
	if err != nil {
		t.Fatal(err)
	}
	if created != 2 || proxies[0].closed.Load() != 0 {
		t.Fatalf("active old adapter closed too early: created=%d closed=%d", created, proxies[0].closed.Load())
	}
	leaseA.release(false)
	if proxies[0].closed.Load() != 0 {
		t.Fatal("old adapter closed while another lease was active")
	}
	leaseB.release(false)
	if proxies[0].closed.Load() != 1 {
		t.Fatalf("retired adapter close count=%d", proxies[0].closed.Load())
	}
	pool.close()
	if proxies[1].closed.Load() != 0 {
		t.Fatal("pool close interrupted an active adapter lease")
	}
	changedLease.release(false)
	if proxies[1].closed.Load() != 1 {
		t.Fatalf("pool close did not close adapter after final lease: %d", proxies[1].closed.Load())
	}
}

func TestAdapterPoolFailedLeaseInvalidatesAdapter(t *testing.T) {
	proxy := &fakeProxy{}
	pool := newAdapterPool(func(models.Node) (constant.Proxy, error) { return proxy, nil })
	_, lease, err := pool.acquire(models.Node{ID: 1, Link: "socks5://node"})
	if err != nil {
		t.Fatal(err)
	}
	lease.release(true)
	if proxy.closed.Load() != 1 {
		t.Fatalf("invalidated adapter close count=%d", proxy.closed.Load())
	}
}

func TestAdapterPoolEvictsLeastRecentlyUsedIdleEntry(t *testing.T) {
	proxies := map[int]*fakeProxy{}
	pool := newAdapterPool(func(node models.Node) (constant.Proxy, error) {
		proxy := &fakeProxy{}
		proxies[node.ID] = proxy
		return proxy, nil
	})
	pool.maxSize = 2
	for _, node := range []models.Node{{ID: 1, Link: "a"}, {ID: 2, Link: "b"}} {
		_, lease, err := pool.acquire(node)
		if err != nil {
			t.Fatal(err)
		}
		lease.release(false)
	}
	_, lease, err := pool.acquire(models.Node{ID: 1, Link: "a"})
	if err != nil {
		t.Fatal(err)
	}
	lease.release(false)
	_, lease, err = pool.acquire(models.Node{ID: 3, Link: "c"})
	if err != nil {
		t.Fatal(err)
	}
	lease.release(false)
	if proxies[2].closed.Load() != 1 || proxies[1].closed.Load() != 0 || proxies[3].closed.Load() != 0 {
		t.Fatalf("unexpected eviction state: node1=%d node2=%d node3=%d", proxies[1].closed.Load(), proxies[2].closed.Load(), proxies[3].closed.Load())
	}
}

func TestAdapterPoolConcurrentAcquireCreatesOnce(t *testing.T) {
	var created atomic.Int32
	pool := newAdapterPool(func(models.Node) (constant.Proxy, error) {
		created.Add(1)
		return &fakeProxy{}, nil
	})
	node := models.Node{ID: 1, Link: "socks5://node"}
	var wg sync.WaitGroup
	wg.Add(20)
	for range 20 {
		go func() {
			defer wg.Done()
			_, lease, err := pool.acquire(node)
			if err != nil {
				t.Errorf("acquire: %v", err)
				return
			}
			lease.release(false)
		}()
	}
	wg.Wait()
	if created.Load() != 1 {
		t.Fatalf("factory calls=%d, want 1", created.Load())
	}
}
