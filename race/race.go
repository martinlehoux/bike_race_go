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
	Organizers            []core.ID
	StartAt               time.Time
	IsOpenForRegistration bool
	RegisteredUsers       []core.ID
}

func NewRace(name string) (Race, error) {
	if len(name) < 3 {
		return Race{}, errors.New("name must be at least 3 characters")
	}
	return Race{
		Id:                    core.NewID(),
		Name:                  name,
		Organizers:            []core.ID{},
		IsOpenForRegistration: false,
	}, nil
}

func LoadRace(ctx context.Context, conn *pgx.Conn, raceId core.ID) (Race, error) {
	var race Race
	err := conn.QueryRow(ctx, `
	SELECT races.id, races.name, races.start_at, races.is_open_for_registration,
	array_agg(race_organizers.user_id) as organizers_ids,
	array_agg(race_registered_users.user_id) filter (where race_registered_users.user_id is not null) as registered_user_ids
	FROM races
	LEFT JOIN race_organizers ON races.id = race_organizers.race_id
	LEFT JOIN race_registered_users ON races.id = race_registered_users.race_id
	WHERE races.id = $1
	GROUP BY races.id, races.name, races.start_at, races.is_open_for_registration
	`, raceId).Scan(&race.Id, &race.Name, &race.StartAt, &race.IsOpenForRegistration, &race.Organizers, &race.RegisteredUsers)
	if err != nil {
		return Race{}, core.Wrap(err, "error selecting races table")
	}
	return race, nil
}

func (race *Race) Save(ctx context.Context, conn *pgx.Conn) error {
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
		`, race.Id, organizer)
		if err != nil {
			return core.Wrap(err, "error upserting race_organizers table")
		}
	}
	for _, userId := range race.RegisteredUsers {
		_, err = tx.Exec(ctx, `
		INSERT INTO race_registered_users (race_id, user_id, registered_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (race_id, user_id) DO NOTHING
		`, race.Id, userId, time.Now())
		if err != nil {
			return core.Wrap(err, "error upserting race_registered_users table")
		}
	}
	return tx.Commit(ctx)
}

func (race *Race) AddOrganizer(user auth.User) error {
	race.Organizers = append(race.Organizers, user.Id)
	return nil
}

func (race *Race) Register(user auth.User) error {
	if core.Find(race.RegisteredUsers, func(userId core.ID) bool { return userId == user.Id }) != nil {
		return errors.New("user already registered")
	}
	if !race.IsOpenForRegistration {
		return errors.New("registration is closed")
	}
	race.RegisteredUsers = append(race.RegisteredUsers, user.Id)
	return nil
}
