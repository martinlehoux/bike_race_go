package config

import (
	"bike_race/core"
	"context"
	"encoding/hex"
	"errors"
	"os"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/exp/slog"
)

var (
	ErrCookieBadLength = errors.New("cookie secret must be 32 bytes")
)

type Config struct {
	Domain       string
	CookieSecret []byte
	DatabaseURL  string
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
		Domain:       domain,
		DatabaseURL:  databaseURL,
		CookieSecret: loadCookieSecret(),
	}
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
		slog.Error(ErrCookieBadLength.Error())
		os.Exit(1)
	}
	slog.Info("cookie secret loaded")
	return cookiesSecret
}

func LoadDatabasePool(ctx context.Context, config Config) *pgxpool.Pool {
	tracer := otelpgx.NewTracer()
	conf, err := pgxpool.ParseConfig(config.DatabaseURL)
	core.Expect(err, "error parsing database url")
	conf.ConnConfig.Tracer = tracer
	pool, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		err = core.Wrap(err, "error connecting to database")
		slog.Error(err.Error())
		os.Exit(1)
	}
	core.Expect(pool.Ping(ctx), "error pinging database")
	slog.Info("connected to database")
	return pool
}
