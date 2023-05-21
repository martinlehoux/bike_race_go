package race

import (
	"bike_race/core"
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

func LoadRace(ctx context.Context, conn *pgx.Conn, raceId core.ID) (Race, error) {
	var race Race
	err := conn.QueryRow(ctx, `
	SELECT
		races.id, races.name, races.start_at, races.is_open_for_registration, races.maximum_participants,
		array_agg(race_organizers.user_id) as organizers_ids,
		array_agg(race_registered_users.user_id) filter (where race_registered_users.user_id is not null) as registered_user_ids
	FROM races
	LEFT JOIN race_organizers ON races.id = race_organizers.race_id
	LEFT JOIN race_registered_users ON races.id = race_registered_users.race_id
	WHERE races.id = $1
	GROUP BY races.id, races.name, races.start_at, races.is_open_for_registration
	`, raceId).Scan(&race.Id, &race.Name, &race.StartAt, &race.IsOpenForRegistration, &race.MaximumParticipants, &race.Organizers, &race.RegisteredUsers)
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
	INSERT INTO races (id, name, start_at, is_open_for_registration, maximum_participants)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (id) DO UPDATE SET name = $2, start_at = $3, is_open_for_registration = $4, maximum_participants = $5
	`, race.Id, race.Name, race.StartAt, race.IsOpenForRegistration, race.MaximumParticipants)
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
