package main

import (
	"bike_race/auth"
	"bike_race/core"
	"bike_race/race"
	"context"
	"encoding/hex"
	"errors"
	"html/template"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"golang.org/x/exp/slog"
)

type IndexTemplateData struct {
	LoggedInUser auth.User
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		err = core.Wrap(err, "error loading .env file")
		slog.Error(err.Error())
		os.Exit(1)
	}
	slog.Info(".env file loaded")
}

func loadCookieSecret() []byte {
	cookiesSecret, err := hex.DecodeString(os.Getenv("COOKIE_SECRET"))
	if err != nil {
		err = core.Wrap(err, "error decoding cookie secret")
		slog.Error(err.Error())
		os.Exit(1)
	}
	if len(cookiesSecret) != 32 {
		err = errors.New("cookie secret must be 32 bytes")
		slog.Error(err.Error())
		os.Exit(1)
	}
	slog.Info("cookie secret loaded")
	return cookiesSecret
}

func connectDatabase() *pgx.Conn {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		err = core.Wrap(err, "error connecting to database")
		slog.Error(err.Error())
		os.Exit(1)
	}
	slog.Info("connected to database")
	return conn
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	ctx := context.Background()
	loadEnv()
	cookiesSecret := loadCookieSecret()
	conn := connectDatabase()
	defer conn.Close(ctx)
	baseTpl := template.Must(template.ParseGlob("templates/base/*.html"))

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(auth.CookieAuthMiddleware(conn, cookiesSecret))

	router.With(middleware.SetHeader("Cache-Control", "max-age=3600")).Handle("/favicon.ico", http.FileServer(http.Dir("static")))

	indexTpl := template.Must(template.Must(baseTpl.Clone()).ParseFiles("templates/index.html"))
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		loggedInUser, _ := auth.UserFromContext(r.Context())
		err := indexTpl.ExecuteTemplate(w, "index.html", IndexTemplateData{LoggedInUser: loggedInUser})
		if err != nil {
			err = core.Wrap(err, "error executing template")
			panic(err)
		}
	})

	router.Mount("/users", auth.Router(conn, baseTpl, cookiesSecret))
	router.Mount("/races", race.Router(conn, baseTpl))

	slog.Info("listening on http://localhost:3000")
	http.ListenAndServe(":3000", router)
}
