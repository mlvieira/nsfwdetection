package middleware

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"

	"github.com/mlvieira/nsfwdetection/internal/config"
)

var (
	once    sync.Once
	limiter *rate.Limiter
)

// ensureLimiter makes sure limiter is only initialized once.
func ensureLimiter() {
	once.Do(func() {
		rps := config.AppConfig.Server.ReqPerSec
		burst := config.AppConfig.Server.Burst
		limiter = rate.NewLimiter(rate.Limit(rps), burst)
	})
}

// RateLimit is a minimal middleware that applies the global limiter to each request.
func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ensureLimiter()

		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
