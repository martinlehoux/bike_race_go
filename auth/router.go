package auth

import (
	"bike_race/config"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/martinlehoux/kagamigo/kauth"
	"github.com/martinlehoux/kagamigo/kcore"
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
		login := LoginFromContext(ctx)
		if !login.Ok {
			kauth.Unauthorized(w, ErrNotAuthenticated)
			return
		}
		users, code, err := UserListQuery(ctx, conn)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}
		page := UsersPage(login, users)
		kcore.RenderPage(ctx, page, w)
	}
}

func viewUserMeRoute() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		login := LoginFromContext(r.Context())
		if !login.Ok {
			kauth.Unauthorized(w, ErrNotAuthenticated)
			return
		}
		page := MePage(login)
		kcore.RenderPage(r.Context(), page, w)
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
			cookie := kauth.CraftCookie(user.Id, config.Auth)
			http.SetCookie(w, &cookie)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}

func logOutRoute() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, ok := kauth.UserFromContext[User](ctx)
		if !ok {
			kauth.Unauthorized(w, ErrNotAuthenticated)
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
	kcore.Expect(err, "error querying user")

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		slog.Warn(ErrBadPassword.Error())
		return User{}, http.StatusBadRequest, ErrBadPassword
	}
	kcore.Expect(err, "error comparing password hash")

	slog.Info("User authenticated", slog.String("username", username))
	return user, http.StatusOK, nil
}
