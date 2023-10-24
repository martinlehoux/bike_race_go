package auth

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/martinlehoux/kagamigo/kcore"
	"golang.org/x/exp/slog"
)

func RegisterUserCommand(ctx context.Context, conn *pgxpool.Pool, username string, password string) (int, error) {
	logger := slog.With(slog.String("command", "RegisterUserCommand"), slog.String("username", username))
	user, err := NewUser(username)
	if err != nil {
		err = kcore.Wrap(err, "error creating user")
		logger.Warn(err.Error())
		return http.StatusBadRequest, err
	}
	logger = logger.With(slog.String("userId", user.Id.String()))
	err = user.SetPassword("", password)
	if err != nil {
		err = kcore.Wrap(err, "error setting password")
		logger.Warn(err.Error())
		return http.StatusBadRequest, err
	}
	err = user.Save(ctx, conn)
	if err != nil {
		err = kcore.Wrap(err, "error saving user")
		logger.Error(err.Error())
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}
