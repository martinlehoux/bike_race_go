package auth

import (
	"bike_race/core"
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/slog"
)

func RegisterUser(ctx context.Context, conn *pgx.Conn, username string, password string) (int, error) {
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
	err = user.Save(conn, ctx)
	if err != nil {
		err = core.Wrap(err, "error saving user")
		logger.Error(err.Error())
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}

func AuthenticateUser(ctx context.Context, conn *pgx.Conn, username string, password string) (User, error) {
	var user User
	err := conn.QueryRow(ctx, `
		SELECT id, username, password_hash
		FROM users
		WHERE username = $1
	`, username).Scan(&user.Id, &user.Username, &user.PasswordHash)
	if err == pgx.ErrNoRows {
		err = errors.New("user not found")
		slog.Warn(err.Error())
		return User{}, err
	} else if err != nil {
		err = core.Wrap(err, "error querying user")
		slog.Error(err.Error())
		return User{}, err
	}
	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		err = errors.New("incorrect password")
		slog.Warn(err.Error())
		return User{}, err
	} else if err != nil {
		err = core.Wrap(err, "error comparing password hash")
		slog.Error(err.Error())
		return User{}, err
	}
	return user, nil
}
