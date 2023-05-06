package race

import (
	"bike_race/auth"
	"bike_race/core"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Race struct {
	Id         core.ID
	Name       string
	Organizers []auth.User
}

func NewRace(name string) (Race, error) {
	if len(name) < 3 {
		return Race{}, errors.New("name must be at least 3 characters")
	}
	return Race{
		Id:         core.NewID(),
		Name:       name,
		Organizers: []auth.User{},
	}, nil
}

func (race *Race) Save(conn *pgx.Conn, ctx context.Context) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	_, err = tx.Exec(ctx, `
	INSERT INTO races (id, name)
	VALUES ($1, $2)
	ON CONFLICT (id) DO UPDATE SET name = $2
	`, race.Id, race.Name)
	if err != nil {
		return fmt.Errorf("error upserting race table: %w", err)
	}
	for _, organizer := range race.Organizers {
		_, err = tx.Exec(ctx, `
		INSERT INTO race_organizers (race_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (race_id, user_id) DO NOTHING
		`, race.Id, organizer.Id)
		if err != nil {
			return fmt.Errorf("error upserting race_organizers table: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (race *Race) AddOrganizer(user auth.User) error {
	race.Organizers = append(race.Organizers, user)
	return nil
}
