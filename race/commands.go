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

func OrganizeRaceCommand(ctx context.Context, conn *pgx.Conn, name string) (int, error) {
	loggedInUser, ok := auth.UserFromContext(ctx)
	if !ok {
		err := errors.New("user not logged in")
		slog.Warn(err.Error())
		return http.StatusUnauthorized, err
	}
	race, err := NewRace(name)
	if err != nil {
		err = core.Wrap(err, "error creating race")
		slog.Warn(err.Error())
		return http.StatusBadRequest, err
	}
	err = race.AddOrganizer(loggedInUser)
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

func OpenRaceForRegistration(ctx context.Context, conn *pgx.Conn, raceId core.ID, maximumParticipants int) (int, error) {
	logger := slog.With(slog.String("raceId", raceId.String()))
	logger.Info("opening race for registration")
	loggedInUser, ok := auth.UserFromContext(ctx)
	if !ok {
		err := errors.New("user not logged in")
		slog.Warn(err.Error())
		return http.StatusUnauthorized, err
	}
	logger = logger.With(slog.String("userId", loggedInUser.Id.String()))
	race, err := LoadRace(ctx, conn, raceId)
	if errors.Is(err, pgx.ErrNoRows) {
		logger.Warn(err.Error())
		return http.StatusNotFound, err
	} else if err != nil {
		panic(err)
	}

	if org := core.Find(race.Organizers, func(userId core.ID) bool { return userId == loggedInUser.Id }); org == nil {
		err = errors.New("user not an organizer")
		logger.Warn(err.Error())
		return http.StatusUnauthorized, err
	}

	err = race.OpenForRegistration(maximumParticipants)
	if err != nil {
		err = core.Wrap(err, "error opening race for registration")
		logger.Warn(err.Error())
		return http.StatusBadRequest, err
	}

	err = race.Save(ctx, conn)
	if err != nil {
		panic(err)
	}

	logger.Info("race opened for registration")
	return http.StatusOK, nil
}

func RegisterForRaceCommand(ctx context.Context, conn *pgx.Conn, raceId core.ID) (int, error) {
	logger := slog.With(slog.String("raceId", raceId.String()))
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		err := errors.New("user not logged in")
		slog.Warn(err.Error())
		return http.StatusUnauthorized, err
	}
	logger = logger.With(slog.String("userId", user.Id.String()))
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
