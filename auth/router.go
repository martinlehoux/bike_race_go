package auth

import (
	"bike_race/core"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/exp/slog"
)

type UsersTemplateData struct {
	LoggedInUser User
	Users        []UserListModel
}

func Router(conn *pgx.Conn, tpl *template.Template, cookiesSecret []byte) chi.Router {
	router := chi.NewRouter()

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		loggedInUser, ok := UserFromContext(ctx)
		if !ok {
			Unauthorized(w, errors.New("not authenticated"))
			return
		}
		users, code, err := UserListQuery(ctx, conn)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}
		templateData := UsersTemplateData{
			LoggedInUser: loggedInUser,
			Users:        users,
		}
		err = tpl.ExecuteTemplate(w, "users.html", templateData)
		if err != nil {
			err = core.Wrap(err, "error executing template")
			slog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	router.Post("/register", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		code, err := RegisterUser(ctx, conn, r.FormValue("username"), r.FormValue("password"))
		if err != nil {
			w.WriteHeader(code)
			w.Write([]byte(err.Error()))
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	})

	router.Post("/log_in", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, err := AuthenticateUser(ctx, conn, r.FormValue("username"), r.FormValue("password"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
	})

	router.Post("/log_out", func(w http.ResponseWriter, r *http.Request) {
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
	})

	return router
}
