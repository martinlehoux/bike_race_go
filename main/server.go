package main

import (
	"bike_race/auth"
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

	router.With(BasicAuthMiddleware(conn)).Get("/users", func(w http.ResponseWriter, r *http.Request) {
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
			log.Printf("error querying users: %v", err)
		}
		defer rows.Close()
		for rows.Next() {
			var user struct{ Username string }
			err := rows.Scan(&user.Username)
			if err != nil {
				log.Printf("error scanning users: %v", err)
			}
			templateData.Users = append(templateData.Users, user)
		}
		err = tpl.ExecuteTemplate(w, "users.html", templateData)
		if err != nil {
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

	http.ListenAndServe(":3000", router)
}
