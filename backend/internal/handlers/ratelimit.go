package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// loginLimiter provides simple per-IP login rate limiting.
// Allows up to maxAttempts failures within window before locking out for lockDur.
type loginLimiter struct {
	mu          sync.Mutex
	attempts    map[string][]time.Time
	maxAttempts int
	window      time.Duration
	lockDur     time.Duration
}

var globalLoginLimiter = &loginLimiter{
	attempts:    make(map[string][]time.Time),
	maxAttempts: 10,
	window:      5 * time.Minute,
	lockDur:     15 * time.Minute,
}

func (l *loginLimiter) clientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		// Take only the first address
		for i, ch := range ip {
			if ch == ',' {
				return ip[:i]
			}
		}
		return ip
	}
	return c.ClientIP()
}

func (l *loginLimiter) record(c *gin.Context) {
	ip := l.clientIP(c)
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	// Prune old entries
	cutoff := now.Add(-l.window)
	filtered := l.attempts[ip][:0]
	for _, t := range l.attempts[ip] {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	l.attempts[ip] = append(filtered, now)
}

func (l *loginLimiter) allow(c *gin.Context) bool {
	ip := l.clientIP(c)
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := now.Add(-l.window)
	var recent []time.Time
	for _, t := range l.attempts[ip] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	l.attempts[ip] = recent
	return len(recent) < l.maxAttempts
}

// LoginRateLimit is a Gin middleware that blocks IPs with too many failed logins.
func LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !globalLoginLimiter.allow(c) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, models429())
			return
		}
		c.Next()
	}
}

func models429() map[string]interface{} {
	return map[string]interface{}{
		"code": 42900,
		"msg":  "登录尝试过于频繁，请稍后再试",
	}
}

// RecordFailedLogin records a failed attempt for rate limiting.
func RecordFailedLogin(c *gin.Context) {
	globalLoginLimiter.record(c)
}
