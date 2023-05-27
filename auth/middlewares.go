package auth

import (
	"bike_race/core"
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/exp/slog"
)

type userContext struct{}

func Unauthorized(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusUnauthorized)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
}

func CookieAuthMiddleware(conn *pgxpool.Pool, secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			cookie, err := r.Cookie("authentication")
			if err == http.ErrNoCookie {
				next.ServeHTTP(w, r)
				return
			} else if err != nil {
				err = core.Wrap(err, "error reading cookie")
				panic(err)
			}
			authentication, err := decrypt(secret, cookie.Value)
			if err != nil {
				err = core.Wrap(err, "error decrypting cookie")
				slog.Warn(err.Error())
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			parts := strings.Split(authentication, ":")
			if len(parts) != 2 {
				err = errors.New("invalid cookie")
				slog.Warn(err.Error())
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			userId, err := core.ParseID(parts[0])
			if err != nil {
				err = core.Wrap(err, "error parsing user id")
				slog.Warn(err.Error())
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			expiresAtSeconds, err := strconv.Atoi(parts[1])
			if err != nil {
				err = core.Wrap(err, "error parsing expires at")
				slog.Warn(err.Error())
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			expiresAt := time.Unix(int64(expiresAtSeconds), 0)
			if time.Now().After(expiresAt) {
				err = errors.New("cookie expired")
				slog.Warn(err.Error())
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			user, err := LoadUser(ctx, conn, userId)
			if err != nil {
				err = core.Wrap(err, "error loading user")
				panic(err)
			}
			ctx = context.WithValue(ctx, userContext{}, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userContext{}).(User)
	return user, ok
}
