package socks5

import (
	"context"
	"fmt"
	"io"
	"net"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"sublink/models"
)

type connectionInfo struct {
	ID            string
	ListenerID    string
	ClientAddress string
	Target        string
	NodeName      string
	ProfileID     string
	ProfileName   string
	Account       string
	Phase         string
	StartedAt     time.Time
	LastActivity  time.Time
	UploadBytes   atomic.Int64
	DownloadBytes atomic.Int64
	cancel        context.CancelFunc
	closeOnce     sync.Once
	mu            sync.Mutex
	client        net.Conn
	upstream      net.Conn
}

func (c *connectionInfo) setRoutingIdentity(identity routingIdentity) {
	c.mu.Lock()
	c.ProfileID = identity.ProfileID
	c.ProfileName = identity.ProfileName
	c.Account = identity.Account
	c.mu.Unlock()
}

func (c *connectionInfo) setUpstream(conn net.Conn, node models.Node) {
	c.mu.Lock()
	c.upstream = conn
	c.NodeName = node.EffectiveName()
	c.Phase = "connected"
	c.LastActivity = time.Now()
	c.mu.Unlock()
}

func (c *connectionInfo) touch() {
	c.mu.Lock()
	c.LastActivity = time.Now()
	c.mu.Unlock()
}

func (c *connectionInfo) close() {
	c.closeOnce.Do(func() {
		if c.cancel != nil {
			c.cancel()
		}
		c.mu.Lock()
		client, upstream := c.client, c.upstream
		c.mu.Unlock()
		if client != nil {
			_ = client.Close()
		}
		if upstream != nil {
			_ = upstream.Close()
		}
	})
}

func (c *connectionInfo) snapshot() ConnectionSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	return ConnectionSnapshot{
		ID: c.ID, ListenerID: c.ListenerID, ClientAddress: c.ClientAddress, Target: c.Target, NodeName: c.NodeName,
		ProfileID: c.ProfileID, ProfileName: c.ProfileName, Account: c.Account,
		Phase: c.Phase, StartedAt: c.StartedAt, LastActivity: c.LastActivity,
		UploadBytes: c.UploadBytes.Load(), DownloadBytes: c.DownloadBytes.Load(),
	}
}

type ConnectionSnapshot struct {
	ID            string    `json:"id"`
	ListenerID    string    `json:"listenerId,omitempty"`
	ClientAddress string    `json:"clientAddress"`
	Target        string    `json:"target"`
	NodeName      string    `json:"nodeName"`
	ProfileID     string    `json:"profileId"`
	ProfileName   string    `json:"profileName"`
	Account       string    `json:"account,omitempty"`
	Phase         string    `json:"phase"`
	StartedAt     time.Time `json:"startedAt"`
	LastActivity  time.Time `json:"lastActivity"`
	UploadBytes   int64     `json:"uploadBytes"`
	DownloadBytes int64     `json:"downloadBytes"`
}

type GatewaySnapshot struct {
	ActiveConnections     int64 `json:"activeConnections"`
	TotalConnections      int64 `json:"totalConnections"`
	SuccessfulConnections int64 `json:"successfulConnections"`
	FailedConnections     int64 `json:"failedConnections"`
	UploadBytes           int64 `json:"uploadBytes"`
	DownloadBytes         int64 `json:"downloadBytes"`
}

type sessionRegistry struct {
	mu         sync.Mutex
	sequence   uint64
	listenerID string
	sessions   map[string]*connectionInfo
	max        int
	perIP      int
	total      atomic.Int64
	success    atomic.Int64
	failed     atomic.Int64
	up         atomic.Int64
	down       atomic.Int64
}

func newSessionRegistry(max, perIP int) *sessionRegistry {
	return &sessionRegistry{sessions: make(map[string]*connectionInfo), max: max, perIP: perIP}
}

func (r *sessionRegistry) setListenerID(listenerID string) {
	r.mu.Lock()
	r.listenerID = strings.TrimSpace(listenerID)
	r.mu.Unlock()
}

func clientIP(address string) string {
	host, _, err := net.SplitHostPort(address)
	if err == nil {
		return host
	}
	return strings.TrimSpace(address)
}

func (r *sessionRegistry) register(parent context.Context, client net.Conn) (*connectionInfo, context.Context, bool) {
	if parent.Err() != nil {
		return nil, nil, false
	}
	ip := clientIP(client.RemoteAddr().String())
	r.mu.Lock()
	defer r.mu.Unlock()
	if parent.Err() != nil {
		return nil, nil, false
	}
	if r.max > 0 && len(r.sessions) >= r.max {
		return nil, nil, false
	}
	if r.perIP > 0 {
		count := 0
		for _, session := range r.sessions {
			if session.ClientAddress == ip {
				count++
			}
		}
		if count >= r.perIP {
			return nil, nil, false
		}
	}
	r.sequence++
	id := fmt.Sprintf("socks5-%d", r.sequence)
	if r.listenerID != "" {
		id = fmt.Sprintf("socks5-%s-%d", r.listenerID, r.sequence)
	}
	ctx, cancel := context.WithCancel(parent)
	now := time.Now()
	session := &connectionInfo{ID: id, ListenerID: r.listenerID, ClientAddress: ip, Phase: "handshake", StartedAt: now, LastActivity: now, cancel: cancel, client: client}
	r.sessions[id] = session
	r.total.Add(1)
	return session, ctx, true
}

func (r *sessionRegistry) remove(id string) {
	r.mu.Lock()
	if session := r.sessions[id]; session != nil {
		delete(r.sessions, id)
		r.up.Add(session.UploadBytes.Load())
		r.down.Add(session.DownloadBytes.Load())
	}
	r.mu.Unlock()
}

func (r *sessionRegistry) snapshot() []ConnectionSnapshot {
	r.mu.Lock()
	items := make([]*connectionInfo, 0, len(r.sessions))
	for _, session := range r.sessions {
		items = append(items, session)
	}
	r.mu.Unlock()
	result := make([]ConnectionSnapshot, 0, len(items))
	for _, session := range items {
		result = append(result, session.snapshot())
	}
	sort.Slice(result, func(i, j int) bool { return result[i].StartedAt.Before(result[j].StartedAt) })
	return result
}

func (r *sessionRegistry) close(id string) bool {
	r.mu.Lock()
	session := r.sessions[id]
	r.mu.Unlock()
	if session == nil {
		return false
	}
	session.close()
	return true
}

func (r *sessionRegistry) closeAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, session := range r.sessions {
		session.close()
	}
}

func (r *sessionRegistry) recordSuccess() { r.success.Add(1) }
func (r *sessionRegistry) recordFailure() { r.failed.Add(1) }
func (r *sessionRegistry) snapshotStats() GatewaySnapshot {
	r.mu.Lock()
	active := int64(len(r.sessions))
	upload, download := r.up.Load(), r.down.Load()
	for _, session := range r.sessions {
		upload += session.UploadBytes.Load()
		download += session.DownloadBytes.Load()
	}
	r.mu.Unlock()
	return GatewaySnapshot{
		ActiveConnections: active, TotalConnections: r.total.Load(),
		SuccessfulConnections: r.success.Load(), FailedConnections: r.failed.Load(),
		UploadBytes: upload, DownloadBytes: download,
	}
}

type trafficConn struct {
	net.Conn
	counter *atomic.Int64
	touch   func()
}

func (c *trafficConn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	if n > 0 {
		c.counter.Add(int64(n))
		c.touch()
	}
	return n, err
}
func (c *trafficConn) Write(p []byte) (int, error) {
	n, err := c.Conn.Write(p)
	if n > 0 {
		c.counter.Add(int64(n))
		c.touch()
	}
	return n, err
}

type activityConn struct {
	net.Conn
	touch func()
}

func (c *activityConn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	if n > 0 {
		c.touch()
	}
	return n, err
}

func (c *activityConn) Write(p []byte) (int, error) {
	n, err := c.Conn.Write(p)
	if n > 0 {
		c.touch()
	}
	return n, err
}

func pumpWithSession(ctx context.Context, session *connectionInfo, client, upstream net.Conn, idleSeconds, maxDurationSeconds int) {
	clientReader := &trafficConn{Conn: client, counter: &session.UploadBytes, touch: session.touch}
	upstreamWriter := &activityConn{Conn: upstream, touch: session.touch}
	upstreamReader := &trafficConn{Conn: upstream, counter: &session.DownloadBytes, touch: session.touch}
	clientWriter := &activityConn{Conn: client, touch: session.touch}

	done := make(chan struct{}, 2)
	go func() { _, _ = io.Copy(upstreamWriter, clientReader); done <- struct{}{} }()
	go func() { _, _ = io.Copy(clientWriter, upstreamReader); done <- struct{}{} }()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	closing := false
	completed := 0
	for completed < 2 {
		if closing {
			completed++
			<-done
			continue
		}
		select {
		case <-done:
			completed++
			if completed == 1 {
				_ = client.Close()
				_ = upstream.Close()
			}
		case <-ctx.Done():
			closing = true
			_ = client.Close()
			_ = upstream.Close()
		case now := <-ticker.C:
			session.mu.Lock()
			lastActivity, started := session.LastActivity, session.StartedAt
			session.mu.Unlock()
			idleExpired := idleSeconds > 0 && now.Sub(lastActivity) >= time.Duration(idleSeconds)*time.Second
			maxExpired := maxDurationSeconds > 0 && now.Sub(started) >= time.Duration(maxDurationSeconds)*time.Second
			if idleExpired || maxExpired {
				closing = true
				_ = client.Close()
				_ = upstream.Close()
			}
		}
	}
}
