package race

import (
	"bike_race/auth"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func OrganizeRace(ctx context.Context, conn *pgx.Conn, name string, user auth.User) (int, error) {
	race, err := NewRace(name)
	if err != nil {
		err = fmt.Errorf("error creating race: %w", err)
		log.Println(err)
		return http.StatusBadRequest, err
	}
	err = race.AddOrganizer(user)
	if err != nil {
		err = fmt.Errorf("error adding organizer: %w", err)
		log.Println(err)
		return http.StatusBadRequest, err
	}
	err = race.Save(conn, ctx)
	if err != nil {
		err = fmt.Errorf("error save race: %w", err)
		log.Println(err)
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}
