package auth

import (
	"bike_race/core"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type UsersTemplateData struct {
	LoggedInUser User
	Users        []struct {
		Username string
	}
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
		templateData := UsersTemplateData{
			LoggedInUser: loggedInUser,
		}
		rows, err := conn.Query(ctx, `SELECT username FROM users`)
		if err != nil {
			err = core.Wrap(err, "error querying users")
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var user struct{ Username string }
			err := rows.Scan(&user.Username)
			if err != nil {
				err = core.Wrap(err, "error scanning users")
				log.Fatal(err)
			}
			templateData.Users = append(templateData.Users, user)
		}
		err = tpl.ExecuteTemplate(w, "users.html", templateData)
		if err != nil {
			err = core.Wrap(err, "error executing template")
			log.Fatal(err)
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
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			expiresAt := time.Now().Add(24 * time.Hour)
			cookieValue, err := encrypt(cookiesSecret, fmt.Sprintf("%s:%d", user.Id.String(), expiresAt.Unix()))
			if err != nil {
				err = core.Wrap(err, "error encrypting cookie")
				log.Fatal(err)
			}
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
