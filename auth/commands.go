package auth

import (
	"bike_race/core"
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"
	"golang.org/x/exp/slog"
)

func RegisterUserCommand(ctx context.Context, conn *pgx.Conn, username string, password string) (int, error) {
	user, err := NewUser(username)
	if err != nil {
		err = core.Wrap(err, "error creating user")
		slog.Warn(err.Error())
		return http.StatusBadRequest, err
	}
	logger := slog.With(slog.String("userId", user.Id.String()))
	err = user.SetPassword("", password)
	if err != nil {
		err = core.Wrap(err, "error setting password")
		logger.Warn(err.Error())
		return http.StatusBadRequest, err
	}
	err = user.Save(ctx, conn)
	if err != nil {
		err = core.Wrap(err, "error saving user")
		logger.Error(err.Error())
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}
