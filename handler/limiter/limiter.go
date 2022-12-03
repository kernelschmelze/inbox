package limiter

import (
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	ips  map[string]*rate.Limiter
	seen map[string]time.Time
	mu   *sync.RWMutex
	r    rate.Limit
	b    int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {

	i := &IPRateLimiter{
		ips:  make(map[string]*rate.Limiter),
		seen: make(map[string]time.Time),
		mu:   &sync.RWMutex{},
		r:    r,
		b:    b,
	}

	return i
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {

	if offset := strings.LastIndex(ip, ":"); offset > 7 {
		ip = ip[:offset]
	}

	i.mu.Lock()

	limiter, exists := i.ips[ip]

	if !exists {

		// housekeeping

		if len(i.ips) > 100 {

			var orphan []string

			for k, v := range i.seen {
				if time.Since(v) > 15*time.Minute {
					orphan = append(orphan, k)
				}
			}

			for k := range orphan {
				delete(i.ips, orphan[k])
				delete(i.seen, orphan[k])
			}

		}

		// new limiter

		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter

	}

	i.seen[ip] = time.Now()

	i.mu.Unlock()

	return limiter
}
