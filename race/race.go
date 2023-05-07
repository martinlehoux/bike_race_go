package race

import (
	"bike_race/auth"
	"bike_race/core"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type Race struct {
	Id                    core.ID
	Name                  string
	Organizers            []auth.User
	StartAt               time.Time
	IsOpenForRegistration bool
}

func NewRace(name string) (Race, error) {
	if len(name) < 3 {
		return Race{}, errors.New("name must be at least 3 characters")
	}
	return Race{
		Id:                    core.NewID(),
		Name:                  name,
		Organizers:            []auth.User{},
		IsOpenForRegistration: false,
	}, nil
}

func (race *Race) Save(conn *pgx.Conn, ctx context.Context) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return core.Wrap(err, "error beginning transaction")
	}
	_, err = tx.Exec(ctx, `
	INSERT INTO races (id, name, start_at, is_open_for_registration)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id) DO UPDATE SET name = $2, start_at = $3, is_open_for_registration = $4
	`, race.Id, race.Name, race.StartAt, race.IsOpenForRegistration)
	if err != nil {
		return core.Wrap(err, "error userting race table")
	}
	for _, organizer := range race.Organizers {
		_, err = tx.Exec(ctx, `
		INSERT INTO race_organizers (race_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (race_id, user_id) DO NOTHING
		`, race.Id, organizer.Id)
		if err != nil {
			return core.Wrap(err, "error upserting race_organizers table")
		}
	}
	return tx.Commit(ctx)
}

func (race *Race) AddOrganizer(user auth.User) error {
	race.Organizers = append(race.Organizers, user)
	return nil
}
