package core

import (
	"net/http"
	"time"

	"golang.org/x/exp/slog"
)

func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		defer func() {
			slog.Info("request finished", slog.String("method", r.Method), slog.String("url", r.URL.String()), slog.Int64("duration_µs", time.Since(startTime).Microseconds()))
		}()
		next.ServeHTTP(w, r)
	})
}
