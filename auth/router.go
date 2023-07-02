package auth

import (
	"bike_race/config"
	"bike_race/core"
	"context"
	"errors"
	"fmt"
	"html/template"
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

type UsersTemplateData struct {
	Users []UserListModel
}

func Router(conn *pgxpool.Pool, baseTpl *template.Template, config config.Config) *chi.Mux {
	router := chi.NewRouter()

	router.Post("/register", registerRoute(conn))
	router.Post("/log_in", logInRoute(conn, config))
	router.Post("/log_out", logOutRoute())

	router.Get("/me", viewUserMeRoute(conn, template.Must(template.Must(baseTpl.Clone()).ParseFiles("templates/me.html"))))
	router.Get("/", viewUsersRoute(conn, template.Must(template.Must(baseTpl.Clone()).ParseFiles("templates/users.html"))))

	return router
}

func viewUsersRoute(conn *pgxpool.Pool, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		data := GetTemplateData(r, UsersTemplateData{})
		if !data.Ok {
			Unauthorized(w, ErrNotAuthenticated)
			return
		}
		users, code, err := UserListQuery(ctx, conn)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}
		data.Data.Users = users
		core.ExecuteTemplate(w, *tpl, "users.html", data)
	}
}

func viewUserMeRoute(conn *pgxpool.Pool, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := GetTemplateData(r, struct{}{})
		if !data.Ok {
			Unauthorized(w, ErrNotAuthenticated)
			return
		}
		core.ExecuteTemplate(w, *tpl, "me.html", data)
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
	if err == pgx.ErrNoRows {
		slog.Warn(ErrUserNotFound.Error())
		return User{}, http.StatusNotFound, ErrUserNotFound
	}
	core.Expect(err, "error querying user")

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		slog.Warn(ErrBadPassword.Error())
		return User{}, http.StatusBadRequest, ErrBadPassword
	}
	core.Expect(err, "error comparing password hash")

	slog.Info("User authenticated", slog.String("username", username))
	return user, http.StatusOK, nil
}
