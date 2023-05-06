package race

import (
	"bike_race/auth"
	"bike_race/core"
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func OrganizeRace(ctx context.Context, conn *pgx.Conn, name string, user auth.User) (int, error) {
	race, err := NewRace(name)
	if err != nil {
		err = core.Wrap(err, "error creating race")
		log.Println(err)
		return http.StatusBadRequest, err
	}
	err = race.AddOrganizer(user)
	if err != nil {
		err = core.Wrap(err, "error adding organizer")
		log.Println(err)
		return http.StatusBadRequest, err
	}
	err = race.Save(conn, ctx)
	if err != nil {
		err = core.Wrap(err, "error save race")
		log.Println(err)
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}
