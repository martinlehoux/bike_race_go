package main

import (
	"bike_race/auth"
	"bike_race/core"
	"bike_race/race"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"html/template"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type IndexTemplateData struct {
	Username string
}

type UsersTemplateData struct {
	Username string
	Users    []struct {
		Username string
	}
}

type RacesTemplateDataRow struct {
	RaceId     core.ID
	RaceName   string
	Organizers string
}

type RacesTemplateData struct {
	Races []RacesTemplateDataRow
}

func unauthorized(w http.ResponseWriter, err error) {
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
				user, err := auth.Authenticate(ctx, conn, username, password)
				if err != nil {
					unauthorized(w, err)
					return
				}
				ctx = context.WithValue(ctx, "user", user)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func main() {
	ctx := context.Background()
	err := godotenv.Load()
	if err != nil {
		err = fmt.Errorf("error loading .env file: %w", err)
		log.Fatal(err)
	}
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	router := chi.NewRouter()
	tpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal(err)
	}

	router.Use(middleware.Logger)
	router.Use(BasicAuthMiddleware(conn))

	router.With(middleware.SetHeader("Cache-Control", "max-age=3600")).Handle("/favicon.ico", http.FileServer(http.Dir("static")))

	router.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, ok := ctx.Value("user").(auth.User)
		if !ok {
			unauthorized(w, errors.New("not authenticated"))
			return
		}
		templateData := UsersTemplateData{
			Username: user.Username,
		}
		rows, err := conn.Query(ctx, `SELECT username FROM users`)
		if err != nil {
			err = fmt.Errorf("error querying users: %w", err)
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var user struct{ Username string }
			err := rows.Scan(&user.Username)
			if err != nil {
				err = fmt.Errorf("error scanning users: %w", err)
				log.Fatal(err)
			}
			templateData.Users = append(templateData.Users, user)
		}
		err = tpl.ExecuteTemplate(w, "users.html", templateData)
		if err != nil {
			err = fmt.Errorf("error executing template: %w", err)
			log.Fatal(err)
		}
	})

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value("user").(auth.User)
		err := tpl.ExecuteTemplate(w, "index.html", IndexTemplateData{Username: user.Username})
		if err != nil {
			log.Fatal(err)
		}
	})

	router.Post("/users/register", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		code, err := auth.RegisterUser(ctx, conn, r.FormValue("username"), r.FormValue("password"))
		if err != nil {
			w.WriteHeader(code)
			w.Write([]byte(err.Error()))
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	})

	router.Get("/races", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, ok := ctx.Value("user").(auth.User)
		if !ok {
			unauthorized(w, errors.New("not authenticated"))
			return
		}
		templateData := RacesTemplateData{}
		rows, err := conn.Query(ctx, `
		SELECT races.id, races.name, string_agg(users.username, ', ')
		FROM races
		LEFT JOIN race_organizers ON races.id = race_organizers.race_id
		LEFT JOIN users ON race_organizers.user_id = users.id
		GROUP BY races.id, races.name
		`)
		if err != nil {
			err = fmt.Errorf("error querying races: %w", err)
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var row RacesTemplateDataRow
			err := rows.Scan(&row.RaceId, &row.RaceName, &row.Organizers)
			if err != nil {
				err = fmt.Errorf("error scanning races: %w", err)
				log.Fatal(err)
			}
			templateData.Races = append(templateData.Races, row)
		}
		err = tpl.ExecuteTemplate(w, "races.html", templateData)
		if err != nil {
			err = fmt.Errorf("error executing template: %w", err)
			log.Fatal(err)
		}
	})

	router.Post("/races/organize", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, ok := ctx.Value("user").(auth.User)
		if !ok {
			unauthorized(w, errors.New("not authenticated"))
			return
		}
		code, err := race.OrganizeRace(ctx, conn, r.FormValue("name"), user)
		if err != nil {
			w.WriteHeader(code)
			w.Write([]byte(err.Error()))
		} else {
			http.Redirect(w, r, "/races", http.StatusSeeOther)
		}
	})

	http.ListenAndServe(":3000", router)
}
