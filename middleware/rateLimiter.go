package middleware

import (
	"net"
	"net/http"
	"time"

	"github.com/fhsmendes/rate-limiter/configs"
	"github.com/fhsmendes/rate-limiter/entity"
	"github.com/fhsmendes/rate-limiter/limiter"
)

var MsgStatus429 = "You have reached the maximum number of requests or actions allowed within a certain time frame\n"

func RateLimitMiddleware(limiter limiter.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("API_KEY")

			limit := configs.Envs.Ip_Rate_Limit
			if limit <= 0 {
				limit = 10
			}

			limitDuration := time.Duration(configs.Envs.Ip_Rate_Limit_Duration_Seconds) * time.Second
			if limitDuration <= 0 {
				limitDuration = 60 * time.Second //1 minute
			}

			blockDuration := time.Duration(configs.Envs.Ip_Block_Duration_Seconds) * time.Second
			if blockDuration <= 0 {
				blockDuration = 300 * time.Second //5 minutes
			}

			if key == "" {
				clientIp, _, _ := net.SplitHostPort(r.RemoteAddr)
				key = clientIp
			} else {
				var tokenData entity.TokenAuth
				if tokenData.Decode(key) != nil {
					http.Error(w, "invalid API key", http.StatusUnauthorized)
					return
				}

				limit = tokenData.RateLimit
				limitDuration = time.Duration(tokenData.RateLimitDuration) * time.Second
				blockDuration = time.Duration(tokenData.BlockDuration) * time.Second
			}

			if key == "" {
				http.Error(w, "API key/IP is required", http.StatusUnauthorized)
				return
			}

			if !limiter.IsAllowed(key, limit, limitDuration, blockDuration) {
				http.Error(w, MsgStatus429, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
