package auth

import (
	"bike_race/core"
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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
				user, err := AuthenticateUser(ctx, conn, username, password)
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

func CookieAuthMiddleware(conn *pgx.Conn, secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			cookie, err := r.Cookie("authentication")
			if err == http.ErrNoCookie {
				next.ServeHTTP(w, r)
				return
			} else if err != nil {
				err = core.Wrap(err, "error reading cookie")
				log.Fatal(err)
			}
			authentication, err := decrypt(secret, cookie.Value)
			if err != nil {
				err = core.Wrap(err, "error decrypting cookie")
				log.Fatal(err)
			}
			parts := strings.Split(authentication, ":")
			if len(parts) != 2 {
				err = errors.New("invalid cookie")
				log.Fatal(err)
			}
			userId, err := core.ParseID(parts[0])
			if err != nil {
				err = core.Wrap(err, "error parsing user id")
				log.Fatal(err)
			}
			expiresAtSeconds, err := strconv.Atoi(parts[1])
			if err != nil {
				err = core.Wrap(err, "error parsing expires at")
				log.Fatal(err)
			}
			expiresAt := time.Unix(int64(expiresAtSeconds), 0)
			if time.Now().After(expiresAt) {
				err = errors.New("cookie expired")
				log.Fatal(err)
			}
			user, err := LoadUser(conn, ctx, userId)
			if err != nil {
				err = core.Wrap(err, "error loading user")
				log.Fatal(err)
			}
			ctx = context.WithValue(ctx, "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
