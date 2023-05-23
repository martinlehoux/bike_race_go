package race

import (
	"bike_race/core"
	"context"

	"github.com/jackc/pgx/v5"
)

func LoadRace(ctx context.Context, conn *pgx.Conn, raceId core.ID) (Race, error) {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return Race{}, core.Wrap(err, "error beginning transaction")
	}
	var race Race
	err = tx.QueryRow(ctx, `
	SELECT
		races.id, races.name, races.start_at, races.is_open_for_registration, races.maximum_participants,
		array_agg(race_organizers.user_id) as organizers_ids
	FROM races
	LEFT JOIN race_organizers ON races.id = race_organizers.race_id
	WHERE races.id = $1
	GROUP BY races.id, races.name, races.start_at, races.is_open_for_registration
	`, raceId).Scan(&race.Id, &race.Name, &race.StartAt, &race.IsOpenForRegistration, &race.MaximumParticipants, &race.Organizers)
	if err != nil {
		return Race{}, core.Wrap(err, "error selecting races table")
	}
	rows, err := tx.Query(ctx, `
	SELECT user_id, registered_at, status
	FROM race_registrations
	WHERE race_id = $1
	`, raceId)
	if err != nil {
		return Race{}, core.Wrap(err, "error selecting race_registrations table")
	}
	defer rows.Close()
	for rows.Next() {
		var registration RaceRegistration
		err := rows.Scan(&registration.UserId, &registration.RegisteredAt, &registration.Status)
		if err != nil {
			return Race{}, core.Wrap(err, "error scanning race_registrations table")
		}
		race.Registrations = append(race.Registrations, registration)
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
	for _, registration := range race.Registrations {
		_, err = tx.Exec(ctx, `
		INSERT INTO race_registrations (race_id, user_id, registered_at, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (race_id, user_id) DO UPDATE SET registered_at = $3, status = $4
		`, race.Id, registration.UserId, registration.RegisteredAt, registration.Status)
		if err != nil {
			return core.Wrap(err, "error upserting race_registrations table")
		}
	}
	return tx.Commit(ctx)
}
