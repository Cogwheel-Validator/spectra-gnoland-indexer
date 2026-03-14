package ratelimit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ValkeyLike abstracts the minimal valkey client behavior used by the rate limiter.
type ValkeyLike interface {
	Increment(key string, ctx context.Context) (int64, error)
	Expirer(key string, ctx context.Context, expiration time.Duration) (bool, error)
}

// KeyStoreLike abstracts the minimal keystore behavior used by the rate limiter.
type KeyStoreLike interface {
	GetKeyLimit(hash [32]byte) (int, bool)
}

type RateLimiter struct {
	valkey         ValkeyLike
	keyStore       KeyStoreLike
	ipRPM          int
	window         time.Duration
	trustedProxies []*net.IPNet
}

func NewRateLimiter(
	vk ValkeyLike,
	ks KeyStoreLike,
	ipRPM int,
	window time.Duration,
	trustedProxyCIDRs []string,
) *RateLimiter {
	var nets []*net.IPNet
	for _, cidr := range trustedProxyCIDRs {
		// Accept bare IPs by normalising them to a /32 or /128 CIDR.
		if !strings.Contains(cidr, "/") {
			if strings.Contains(cidr, ":") {
				cidr += "/128"
			} else {
				cidr += "/32"
			}
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			log.Printf("ratelimit: ignoring invalid trusted proxy CIDR %q: %v", cidr, err)
			continue
		}
		nets = append(nets, ipNet)
	}
	return &RateLimiter{
		valkey:         vk,
		keyStore:       ks,
		ipRPM:          ipRPM,
		window:         window,
		trustedProxies: nets,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")

		var identifier string
		var limit int

		if apiKey != "" {
			hash := sha256.Sum256([]byte(apiKey))
			rpmLimit, found := rl.keyStore.GetKeyLimit(hash)
			if !found {
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}
			identifier = "key:" + hex.EncodeToString(hash[:8])
			limit = rpmLimit
		} else {
			ip := rl.realIP(r)
			identifier = "ip:" + ip
			limit = rl.ipRPM
		}

		valkeyKey := fmt.Sprintf("rl:%s", identifier)
		count, err := rl.valkey.Increment(valkeyKey, r.Context())
		if err != nil {
			log.Printf("ratelimit: valkey error (fail-open): %v", err)
			next.ServeHTTP(w, r)
			return
		}

		if count == 1 {
			if _, err := rl.valkey.Expirer(valkeyKey, r.Context(), rl.window); err != nil {
				log.Printf("ratelimit: valkey expire error: %v", err)
			}
		}

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
		remaining := max(limit-int(count), 0)
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

		if int(count) > limit {
			w.Header().Set("Retry-After", strconv.Itoa(int(rl.window.Seconds())))
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// realIP returns the best-effort client IP for the request.
// Forwarded headers (X-Forwarded-For, X-Real-Ip) are only trusted when
// RemoteAddr belongs to a configured trusted proxy CIDR; otherwise RemoteAddr
// is used directly to prevent header-spoofing bypasses.
func (rl *RateLimiter) realIP(r *http.Request) string {
	remoteHost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteHost = r.RemoteAddr
	}

	if len(rl.trustedProxies) > 0 {
		remoteIP := net.ParseIP(remoteHost)
		if remoteIP != nil && rl.isTrustedProxy(remoteIP) {
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				if i := strings.IndexByte(xff, ','); i > 0 {
					return strings.TrimSpace(xff[:i])
				}
				return strings.TrimSpace(xff)
			}
			if xrip := r.Header.Get("X-Real-Ip"); xrip != "" {
				return strings.TrimSpace(xrip)
			}
		}
	}

	return remoteHost
}

func (rl *RateLimiter) isTrustedProxy(ip net.IP) bool {
	for _, network := range rl.trustedProxies {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}
