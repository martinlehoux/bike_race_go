package auth

import (
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

type UsersTemplateData struct {
	Users []UserListModel
}

func Router(conn *pgxpool.Pool, baseTpl *template.Template, cookiesSecret []byte) chi.Router {
	router := chi.NewRouter()

	router.Post("/register", registerRoute(conn))
	router.Post("/log_in", logInRoute(conn, cookiesSecret))
	router.Post("/log_out", logOutRoute())

	router.Get("/", viewUsersRoute(conn, template.Must(template.Must(baseTpl.Clone()).ParseFiles("templates/users.html"))))

	return router
}

func viewUsersRoute(conn *pgxpool.Pool, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		data := GetTemplateData(r, UsersTemplateData{})
		if !data.Ok {
			Unauthorized(w, errors.New("not authenticated"))
			return
		}
		users, code, err := UserListQuery(ctx, conn)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}
		data.Data.Users = users
		err = tpl.ExecuteTemplate(w, "users.html", data)
		if err != nil {
			err = core.Wrap(err, "error executing template")
			slog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

func logInRoute(conn *pgxpool.Pool, cookiesSecret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, code, err := AuthenticateUser(ctx, conn, r.FormValue("username"), r.FormValue("password"))
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		} else {
			expiresAt := time.Now().Add(24 * time.Hour)
			cookieValue := encrypt(cookiesSecret, fmt.Sprintf("%s:%d", user.Id.String(), expiresAt.Unix()))
			http.SetCookie(w, &http.Cookie{
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
			Unauthorized(w, errors.New("not authenticated"))
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
		err = errors.New("user not found")
		slog.Warn(err.Error())
		return User{}, http.StatusNotFound, err
	} else if err != nil {
		err = core.Wrap(err, "error querying user")
		panic(err)
	}
	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		err = errors.New("incorrect password")
		slog.Warn(err.Error())
		return User{}, http.StatusBadRequest, err
	} else if err != nil {
		err = core.Wrap(err, "error comparing password hash")
		panic(err)
	}

	slog.Info("User authenticated", slog.String("username", username))
	return user, http.StatusOK, nil
}
