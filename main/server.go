package main

import (
	"bike_race/auth"
	"bike_race/core"
	"bike_race/race"
	"context"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type IndexTemplateData struct {
	Username string
}

func main() {
	ctx := context.Background()
	err := godotenv.Load()
	if err != nil {
		err = core.Wrap(err, "error loading .env file")
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
	router.Use(auth.BasicAuthMiddleware(conn))

	router.With(middleware.SetHeader("Cache-Control", "max-age=3600")).Handle("/favicon.ico", http.FileServer(http.Dir("static")))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value("user").(auth.User)
		err := tpl.ExecuteTemplate(w, "index.html", IndexTemplateData{Username: user.Username})
		if err != nil {
			log.Fatal(err)
		}
	})

	router.Mount("/users", auth.Router(conn, tpl))
	router.Mount("/races", race.Router(conn, tpl))

	http.ListenAndServe(":3000", router)
}
