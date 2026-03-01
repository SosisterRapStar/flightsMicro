package middleware

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/config"
)

type MiddlewareFunc func(http.Handler) http.Handler

type Middleware struct {
	Logger     MiddlewareFunc
	Monitoring MiddlewareFunc
	Timeout    MiddlewareFunc
}

func NewMiddleware(
	cfg *config.AppConfig,
) Middleware {
	return Middleware{
		Logger: Logger,
	}
}

func Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			sw := newStatusWriter(w)

			h.ServeHTTP(sw, r)

			entry := log.WithFields(log.Fields{
				"method":   r.Method,
				"remote":   r.RemoteAddr,
				"uri":      r.RequestURI,
				"proto":    r.Proto,
				"duration": time.Since(startTime),
				"status":   sw.Status(),
			})

			switch {
			case sw.Status() >= 500:
				entry.Warn("request failed due to internal error")
			case sw.Status() >= 400:
				entry.Info("request completed with client error")
			default:
				entry.Info("request completed")
			}
		},
	)
}
