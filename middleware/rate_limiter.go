package middleware

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// ------------------------- CONFIG -------------------------
type AlgorithmType string

const (
	FixedWindowAlgo   AlgorithmType = "fixed_window"
	SlidingWindowAlgo AlgorithmType = "sliding_window"
	TokenBucketAlgo   AlgorithmType = "token_bucket"
)

type RateLimiterConfig struct {
	Algorithm AlgorithmType
	Requests  int           // allowed requests
	Window    time.Duration // time window
	BucketCap int           // for token bucket
	Refill    int           // refill rate per second
	Redis     *redis.Client // Redis connection
}

// ------------------------- INTERFACE -------------------------
type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

// ------------------------- FACTORY -------------------------
func NewRateLimiter(cfg RateLimiterConfig) (RateLimiter, error) {
	if cfg.Redis == nil {
		return nil, errors.New("Redis client required")
	}

	switch cfg.Algorithm {
	case FixedWindowAlgo:
		return &FixedWindowLimiter{cfg}, nil
	case SlidingWindowAlgo:
		return &SlidingWindowLimiter{cfg}, nil
	case TokenBucketAlgo:
		return &TokenBucketLimiter{cfg}, nil
	default:
		return nil, errors.New("unsupported rate limiting algorithm")
	}
}

// ------------------------- FIXED WINDOW -------------------------
type FixedWindowLimiter struct {
	cfg RateLimiterConfig
}

func (f *FixedWindowLimiter) Allow(ctx context.Context, key string) (bool, error) {
	rdb := f.cfg.Redis
	windowKey := "fw:" + key
	count, err := rdb.Incr(ctx, windowKey).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		rdb.Expire(ctx, windowKey, f.cfg.Window)
	}

	return count <= int64(f.cfg.Requests), nil
}

// ------------------------- SLIDING WINDOW -------------------------
type SlidingWindowLimiter struct {
	cfg RateLimiterConfig
}

func (s *SlidingWindowLimiter) Allow(ctx context.Context, key string) (bool, error) {
	rdb := s.cfg.Redis
	now := time.Now().UnixNano()
	windowStart := now - s.cfg.Window.Nanoseconds()
	listKey := "sw:" + key

	// remove old timestamps
	rdb.ZRemRangeByScore(ctx, listKey, "0", strconv.FormatInt(windowStart, 10))

	// count active timestamps
	count, _ := rdb.ZCard(ctx, listKey).Result()
	if count >= int64(s.cfg.Requests) {
		return false, nil
	}

	// add new timestamp
	rdb.ZAdd(ctx, listKey, redis.Z{Score: float64(now), Member: now})
	rdb.Expire(ctx, listKey, s.cfg.Window)
	return true, nil
}

// ------------------------- TOKEN BUCKET -------------------------
type TokenBucketLimiter struct {
	cfg RateLimiterConfig
}

func (t *TokenBucketLimiter) Allow(ctx context.Context, key string) (bool, error) {
	rdb := t.cfg.Redis
	bucketKey := "tb:" + key

	lastRefillKey := bucketKey + ":ts"
	now := time.Now().Unix()

	lastRefillStr, _ := rdb.Get(ctx, lastRefillKey).Result()
	lastRefill, _ := strconv.ParseInt(lastRefillStr, 10, 64)

	elapsed := now - lastRefill
	newTokens := int64(elapsed) * int64(t.cfg.Refill)
	currTokens, _ := rdb.Get(ctx, bucketKey).Int64()

	currTokens = min(int64(t.cfg.BucketCap), currTokens+newTokens)
	if currTokens <= 0 {
		return false, nil
	}

	rdb.Set(ctx, bucketKey, currTokens-1, 0)
	rdb.Set(ctx, lastRefillKey, now, 0)
	return true, nil
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func Middleware(cfg RateLimiterConfig) func(http.Handler) http.Handler {
	rl, err := NewRateLimiter(cfg)
	if err != nil {
		panic(err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.Background()

			// Identify user from header or IP
			userID := r.Header.Get("X-User-ID")
			if userID == "" {
				userID = clientIP(r)
			}

			key := "ratelimit:" + userID
			allowed, err := rl.Allow(ctx, key)
			if err != nil {
				http.Error(w, "Rate limiter error", http.StatusInternalServerError)
				return
			}

			if !allowed {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("Too Many Requests"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}
	return ip
}
