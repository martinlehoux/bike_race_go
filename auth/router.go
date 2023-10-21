package auth

import (
	"bike_race/config"
	"bike_race/core"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/slog"
)

var (
	ErrNotAuthenticated = errors.New("not authenticated")
)

func Router(conn *pgxpool.Pool, config config.Config) *chi.Mux {
	router := chi.NewRouter()

	router.Post("/register", registerRoute(conn))
	router.Post("/log_in", logInRoute(conn, config))
	router.Post("/log_out", logOutRoute())

	router.Get("/me", viewUserMeRoute())
	router.Get("/", viewUsersRoute(conn))

	return router
}

func viewUsersRoute(conn *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		lc := GetLoginContext(r)
		if !lc.LoggedIn {
			Unauthorized(w, ErrNotAuthenticated)
			return
		}
		users, code, err := UserListQuery(ctx, conn)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}
		page := UsersPage(users, lc)
		core.RenderPage(ctx, page, w)
	}
}

func viewUserMeRoute() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lc := GetLoginContext(r)
		if !lc.LoggedIn {
			Unauthorized(w, ErrNotAuthenticated)
			return
		}
		page := MePage(lc)
		core.RenderPage(r.Context(), page, w)
	}
}

func registerRoute(conn *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		code, err := RegisterUserCommand(ctx, conn, r.FormValue("username"), r.FormValue("password"))
		if err != nil {
			http.Error(w, err.Error(), code)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}

func logInRoute(conn *pgxpool.Pool, config config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, code, err := AuthenticateUser(ctx, conn, r.FormValue("username"), r.FormValue("password"))
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		} else {
			expiresAt := time.Now().Add(24 * time.Hour)
			cookieValue := encrypt(config.CookieSecret, fmt.Sprintf("%s:%d", user.Id.String(), expiresAt.Unix()))
			http.SetCookie(w, &http.Cookie{
				Domain:  config.Domain,
				Name:    "authentication",
				Value:   cookieValue,
				Expires: expiresAt,
				Path:    "/",
			})
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}

func logOutRoute() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, ok := UserFromContext(ctx)
		if !ok {
			Unauthorized(w, ErrNotAuthenticated)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:    "authentication",
			Value:   "",
			Expires: time.Unix(0, 0),
			Path:    "/",
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func AuthenticateUser(ctx context.Context, conn *pgxpool.Pool, username string, password string) (User, int, error) {
	slog.Info("Authenticating user", slog.String("username", username))
	var user User
	err := conn.QueryRow(ctx, `
		SELECT id, username, password_hash
		FROM users
		WHERE username = $1
	`, username).Scan(&user.Id, &user.Username, &user.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		slog.Warn(ErrUserNotFound.Error())
		return User{}, http.StatusNotFound, ErrUserNotFound
	}
	core.Expect(err, "error querying user")

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		slog.Warn(ErrBadPassword.Error())
		return User{}, http.StatusBadRequest, ErrBadPassword
	}
	core.Expect(err, "error comparing password hash")

	slog.Info("User authenticated", slog.String("username", username))
	return user, http.StatusOK, nil
}
