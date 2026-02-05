package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
)

type RateLimiter struct {
	redis          *redis.Client
	limiter        *redis_rate.Limiter
	enabled        bool
	circuitBreaker *gobreaker.CircuitBreaker
	trustedProxies []net.IPNet
	securityLogger *SecurityLogger
}

type RateLimitStrategy func(r *http.Request) string

type RateLimitRule struct {
	Limit    int
	Window   time.Duration
	Burst    int
	Strategy RateLimitStrategy
}

func NewRateLimiter(redisClient *redis.Client, enabled bool, trustedProxyCIDRs []string, secLogger *SecurityLogger) *RateLimiter {
	trustedProxies := parseTrustedProxies(trustedProxyCIDRs)

	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "redis-rate-limiter",
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
	})

	return &RateLimiter{
		redis:          redisClient,
		limiter:        redis_rate.NewLimiter(redisClient),
		enabled:        enabled,
		circuitBreaker: cb,
		trustedProxies: trustedProxies,
		securityLogger: secLogger,
	}
}

func parseTrustedProxies(cidrs []string) []net.IPNet {
	var proxies []net.IPNet
	for _, cidr := range cidrs {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err == nil {
			proxies = append(proxies, *ipnet)
		}
	}
	return proxies
}

func (rl *RateLimiter) isTrustedProxy(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, ipnet := range rl.trustedProxies {
		if ipnet.Contains(parsedIP) {
			return true
		}
	}
	return false
}

func ByIP(rl *RateLimiter) RateLimitStrategy {
	return func(r *http.Request) string {
		remoteIP := parseIP(r.RemoteAddr)

		if rl.isTrustedProxy(remoteIP) {
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				clientIP := strings.Split(xff, ",")[0]
				clientIP = strings.TrimSpace(clientIP)
				return clientIP
			}
		} else if r.Header.Get("X-Forwarded-For") != "" {
			if rl.securityLogger != nil {
				rl.securityLogger.LogIPSpoofing(r, r.Header.Get("X-Forwarded-For"))
			}
		}

		return remoteIP
	}
}

func parseIP(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

func ByUser(rl *RateLimiter) RateLimitStrategy {
	return func(r *http.Request) string {
		if userID := r.Context().Value("user_id"); userID != nil {
			return fmt.Sprintf("user:%v", userID)
		}
		return ByIP(rl)(r)
	}
}

func ByRoute(route string, rl *RateLimiter) RateLimitStrategy {
	return func(r *http.Request) string {
		return fmt.Sprintf("route:%s:%s", route, ByIP(rl)(r))
	}
}

func Composite(strategies ...RateLimitStrategy) RateLimitStrategy {
	return func(r *http.Request) string {
		key := ""
		for i, strategy := range strategies {
			if i > 0 {
				key += ":"
			}
			key += strategy(r)
		}
		return key
	}
}

func (rl *RateLimiter) Limit(rule RateLimitRule) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.enabled || rl.redis == nil {
				next.ServeHTTP(w, r)
				return
			}

			key := rule.Strategy(r)
			if key == "" {
				key = "default"
			}
			key = fmt.Sprintf("ratelimit:%s", key)

			limit := redis_rate.Limit{
				Rate:   rule.Limit,
				Period: rule.Window,
				Burst:  rule.Burst,
			}

			result, err := rl.circuitBreaker.Execute(func() (interface{}, error) {
				return rl.limiter.Allow(r.Context(), key, limit)
			})
			if err != nil {
				if rl.securityLogger != nil {
					rl.securityLogger.LogEvent(
						"circuit_breaker_open",
						SeverityHigh,
						r,
						map[string]string{"error": err.Error()},
					)
				}
				http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
				return
			}

			res := result.(*redis_rate.Result)
			resetTime := time.Now().Add(res.ResetAfter)
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rule.Limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

			if res.Allowed == 0 {
				w.Header().Set("Retry-After", strconv.Itoa(int(res.RetryAfter.Seconds())))

				if rl.securityLogger != nil {
					rl.securityLogger.LogRateLimitExceeded(r, rule.Limit, rule.Window.String())
				}

				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) GlobalLimit(limit int, window time.Duration, burst int) func(next http.Handler) http.Handler {
	return rl.Limit(RateLimitRule{
		Limit:    limit,
		Window:   window,
		Burst:    burst,
		Strategy: ByIP(rl),
	})
}

func (rl *RateLimiter) PerUserLimit(limit int, window time.Duration, burst int) func(next http.Handler) http.Handler {
	return rl.Limit(RateLimitRule{
		Limit:    limit,
		Window:   window,
		Burst:    burst,
		Strategy: ByUser(rl),
	})
}

func (rl *RateLimiter) PerRouteLimit(route string, limit int, window time.Duration, burst int) func(next http.Handler) http.Handler {
	return rl.Limit(RateLimitRule{
		Limit:    limit,
		Window:   window,
		Burst:    burst,
		Strategy: ByRoute(route, rl),
	})
}
