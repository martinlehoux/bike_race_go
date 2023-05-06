package auth

import (
	"bike_race/core"
	"errors"
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type UsersTemplateData struct {
	Username string
	Users    []struct {
		Username string
	}
}

func Router(conn *pgx.Conn, tpl *template.Template) chi.Router {
	router := chi.NewRouter()

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, ok := ctx.Value("user").(User)
		if !ok {
			Unauthorized(w, errors.New("not authenticated"))
			return
		}
		templateData := UsersTemplateData{
			Username: user.Username,
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

	return router
}
