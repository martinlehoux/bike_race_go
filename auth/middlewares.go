package auth

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func Unauthorized(w http.ResponseWriter, err error) {
	w.Header().Add("WWW-Authenticate", `Basic realm="private"`)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(err.Error()))
}

func BasicAuthMiddleware(conn *pgx.Conn) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			username, password, ok := r.BasicAuth()
			if ok {
				user, err := Authenticate(ctx, conn, username, password)
				if err != nil {
					Unauthorized(w, err)
					return
				}
				ctx = context.WithValue(ctx, "user", user)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
