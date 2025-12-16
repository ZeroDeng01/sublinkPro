package api

import (
	"sublink/models"
	"sublink/utils"
	"sync"
	"time"
)

type AttemptInfo struct {
	Count         int       // 失败次数
	FirstFailTime time.Time // 首次失败时间
	BanUntil      time.Time // 封禁截止时间
}

type LimitManager struct {
	mu       sync.Mutex
	attempts map[string]*AttemptInfo
}

var (
	loginLimiter *LimitManager
	limiterOnce  sync.Once
)

// GetLoginLimiter 获取单例限流器
func GetLoginLimiter() *LimitManager {
	limiterOnce.Do(func() {
		loginLimiter = &LimitManager{
			attempts: make(map[string]*AttemptInfo),
		}
		// 启动清理任务
		go loginLimiter.cleanupLoop()
	})
	return loginLimiter
}

// CheckBan 检查是否被封禁
// 返回: isBanned, banUntil
func (m *LimitManager) CheckBan(ip string) (bool, time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, exists := m.attempts[ip]
	if !exists {
		return false, time.Time{}
	}

	// 检查是否在封禁期内
	if !info.BanUntil.IsZero() && time.Now().Before(info.BanUntil) {
		return true, info.BanUntil
	}

	// 如果封禁时间已过，清除封禁状态（保留计数方便后续重新计算窗口，或者直接重置也可以，这里选择lazy clean，或者让RecordFailure处理）
	if !info.BanUntil.IsZero() && time.Now().After(info.BanUntil) {
		delete(m.attempts, ip) // 封禁过期，重置
		return false, time.Time{}
	}

	return false, time.Time{}
}

// RecordFailure 记录失败
func (m *LimitManager) RecordFailure(ip string) {
	config := models.ReadConfig()
	failCountLimit := config.LoginFailCount
	failWindow := time.Duration(config.LoginFailWindow) * time.Minute
	banDuration := time.Duration(config.LoginBanDuration) * time.Minute

	m.mu.Lock()
	defer m.mu.Unlock()

	info, exists := m.attempts[ip]
	now := time.Now()

	if !exists {
		m.attempts[ip] = &AttemptInfo{
			Count:         1,
			FirstFailTime: now,
		}
		// 如果配置允许1次就封禁（虽然不常见）
		if failCountLimit <= 1 {
			m.attempts[ip].BanUntil = now.Add(banDuration)
			utils.Warn("IP %s 因登录失败被封禁至 %v", ip, m.attempts[ip].BanUntil)
		}
		return
	}

	// 如果已经在封禁中，直接返回
	if !info.BanUntil.IsZero() && now.Before(info.BanUntil) {
		return
	}

	// 检查是否在统计窗口内
	if now.Sub(info.FirstFailTime) > failWindow {
		// 超过窗口期，重置计数
		info.Count = 1
		info.FirstFailTime = now
		info.BanUntil = time.Time{} // 确保清除封禁标记
	} else {
		// 在窗口期内，增加计数
		info.Count++
		if info.Count >= failCountLimit {
			info.BanUntil = now.Add(banDuration)
			utils.Warn("IP %s 因 %d 分钟内失败 %d 次被封禁至 %v", ip, config.LoginFailWindow, info.Count, info.BanUntil)
		}
	}
}

// ClearFailures 清除失败记录（登录成功后调用）
func (m *LimitManager) ClearFailures(ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.attempts, ip)
}

// cleanupLoop 定期清理过期的记录防止内存泄漏
func (m *LimitManager) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		config := models.ReadConfig()
		// 使用稍微宽裕一点的过期判断
		window := time.Duration(config.LoginFailWindow)*time.Minute + time.Minute

		for ip, info := range m.attempts {
			// 如果有封禁且已过期，删除
			if !info.BanUntil.IsZero() && now.After(info.BanUntil) {
				delete(m.attempts, ip)
				continue
			}
			// 如果没有封禁，但早已超过窗口期，删除
			if info.BanUntil.IsZero() && now.Sub(info.FirstFailTime) > window {
				delete(m.attempts, ip)
			}
		}
		m.mu.Unlock()
	}
}
