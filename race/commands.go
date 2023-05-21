package race

import (
	"bike_race/auth"
	"bike_race/core"
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"golang.org/x/exp/slog"
)

func OrganizeRaceCommand(ctx context.Context, conn *pgx.Conn, name string, user auth.User) (int, error) {
	race, err := NewRace(name)
	if err != nil {
		err = core.Wrap(err, "error creating race")
		slog.Warn(err.Error())
		return http.StatusBadRequest, err
	}
	err = race.AddOrganizer(user)
	if err != nil {
		err = core.Wrap(err, "error adding organizer")
		slog.Warn(err.Error())
		return http.StatusBadRequest, err
	}
	err = race.Save(ctx, conn)
	if err != nil {
		err = core.Wrap(err, "error saving race")
		slog.Error(err.Error())
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}

func RegisterForRaceCommand(ctx context.Context, conn *pgx.Conn, raceId core.ID, user auth.User) (int, error) {
	logger := slog.With(slog.String("userId", user.Id.String()), slog.String("raceId", raceId.String()))

	race, err := LoadRace(ctx, conn, raceId)
	if errors.Is(err, pgx.ErrNoRows) {
		logger.Warn(err.Error())
		return http.StatusNotFound, err
	} else if err != nil {
		panic(err)
	}

	err = race.Register(user)
	if err != nil {
		err = core.Wrap(err, "error registering user")
		logger.Warn(err.Error())
		return http.StatusBadRequest, err
	}

	err = race.Save(ctx, conn)
	if err != nil {
		panic(err)
	}

	logger.Info("user registered to race")
	return http.StatusOK, nil
}
