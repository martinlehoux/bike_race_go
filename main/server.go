package main

import (
	"bike_race/auth"
	"bike_race/core"
	"bike_race/race"
	"context"
	"encoding/hex"
	"errors"
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
	LoggedInUser auth.User
}

func main() {
	ctx := context.Background()
	err := godotenv.Load()
	if err != nil {
		err = core.Wrap(err, "error loading .env file")
		log.Fatal(err)
	}
	cookiesSecret, err := hex.DecodeString(os.Getenv("COOKIE_SECRET"))
	if err != nil {
		err = core.Wrap(err, "error decoding cookie secret")
		log.Fatal(err)
	}
	if len(cookiesSecret) != 32 {
		err = errors.New("cookie secret must be 32 bytes")
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
	router.Use(middleware.Recoverer)
	router.Use(auth.CookieAuthMiddleware(conn, cookiesSecret))

	router.With(middleware.SetHeader("Cache-Control", "max-age=3600")).Handle("/favicon.ico", http.FileServer(http.Dir("static")))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		loggedInUser, _ := auth.UserFromContext(r.Context())
		err := tpl.ExecuteTemplate(w, "index.html", IndexTemplateData{LoggedInUser: loggedInUser})
		if err != nil {
			log.Fatal(err)
		}
	})

	router.Mount("/users", auth.Router(conn, tpl, cookiesSecret))
	router.Mount("/races", race.Router(conn, tpl))

	http.ListenAndServe(":3000", router)
}
