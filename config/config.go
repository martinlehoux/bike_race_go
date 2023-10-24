package config

import (
	"context"
	"errors"
	"os"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/martinlehoux/kagamigo/kauth"
	"github.com/martinlehoux/kagamigo/kcore"
	"golang.org/x/exp/slog"
)

var (
	ErrCookieBadLength = errors.New("cookie secret must be 32 bytes")
)

type Config struct {
	DatabaseURL string
	Auth        kauth.AuthConfig
}

func LoadConfig() Config {
	loadEnv()
	domain := os.Getenv("DOMAIN")
	if domain == "" {
		slog.Error("DOMAIN environment variable is required")
		os.Exit(1)
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		slog.Error("DATABASE_URL environment variable is required")
		os.Exit(1)
	}
	return Config{
		DatabaseURL: databaseURL,
		Auth: kauth.AuthConfig{
			Domain:       domain,
			CookieSecret: kauth.LoadCookieSecret(os.Getenv("COOKIE_SECRET")),
		},
	}
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		err = kcore.Wrap(err, "error loading .env file")
		slog.Error(err.Error())
		os.Exit(1)
	}
	slog.Info(".env file loaded")
}

func LoadDatabasePool(ctx context.Context, config Config) *pgxpool.Pool {
	tracer := otelpgx.NewTracer()
	conf, err := pgxpool.ParseConfig(config.DatabaseURL)
	kcore.Expect(err, "error parsing database url")
	conf.ConnConfig.Tracer = tracer
	pool, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		err = kcore.Wrap(err, "error connecting to database")
		slog.Error(err.Error())
		os.Exit(1)
	}
	kcore.Expect(pool.Ping(ctx), "error pinging database")
	slog.Info("connected to database")
	return pool
}
