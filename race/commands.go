package race

import (
	"bike_race/auth"
	"bike_race/core"
	"context"
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
	err = race.Save(conn, ctx)
	if err != nil {
		err = core.Wrap(err, "error saving race")
		slog.Error(err.Error())
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}
