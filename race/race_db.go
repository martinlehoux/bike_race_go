package race

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/martinlehoux/kagamigo/kcore"
)

func LoadRace(ctx context.Context, conn *pgxpool.Pool, raceId kcore.ID) (Race, error) {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return Race{}, kcore.Wrap(err, "error beginning transaction")
	}
	var race Race
	err = tx.QueryRow(ctx, `
	SELECT
		races.id, races.name, races.start_at, races.is_open_for_registration, races.maximum_participants, races.cover_image_id,
		array_agg(race_organizers.user_id) as organizers_ids
	FROM races
	LEFT JOIN race_organizers ON races.id = race_organizers.race_id
	WHERE races.id = $1
	GROUP BY races.id, races.name, races.start_at, races.is_open_for_registration
	`, raceId).Scan(&race.Id, &race.Name, &race.StartAt, &race.IsOpenForRegistration, &race.MaximumParticipants, &race.CoverImage, &race.Organizers)
	if err != nil {
		return Race{}, kcore.Wrap(err, "error selecting races table")
	}
	race.Registrations = map[kcore.ID]RaceRegistration{}
	rows, err := tx.Query(ctx, `
	SELECT user_id, registered_at, status, medical_certificate, is_medical_certificate_approved
	FROM race_registrations
	WHERE race_id = $1
	`, raceId)
	if err != nil {
		return Race{}, kcore.Wrap(err, "error selecting race_registrations table")
	}
	defer rows.Close()
	for rows.Next() {
		var registration RaceRegistration
		err := rows.Scan(&registration.UserId, &registration.RegisteredAt, &registration.Status, &registration.MedicalCertificate, &registration.IsMedicalCertificateApproved)
		if err != nil {
			return Race{}, kcore.Wrap(err, "error scanning race_registrations table")
		}
		race.Registrations[registration.UserId] = registration
	}
	return race, nil
}

func (race *Race) Save(ctx context.Context, conn *pgxpool.Pool) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return kcore.Wrap(err, "error beginning transaction")
	}
	_, err = tx.Exec(ctx, `
	INSERT INTO races (id, name, start_at, is_open_for_registration, maximum_participants, cover_image_id)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (id) DO UPDATE SET name = $2, start_at = $3, is_open_for_registration = $4, maximum_participants = $5, cover_image_id = $6
	`, race.Id, race.Name, race.StartAt, race.IsOpenForRegistration, race.MaximumParticipants, race.CoverImage)
	if err != nil {
		return kcore.Wrap(err, "error userting race table")
	}
	for _, organizer := range race.Organizers {
		_, err = tx.Exec(ctx, `
		INSERT INTO race_organizers (race_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (race_id, user_id) DO NOTHING
		`, race.Id, organizer)
		if err != nil {
			return kcore.Wrap(err, "error upserting race_organizers table")
		}
	}
	for _, registration := range race.Registrations {
		_, err = tx.Exec(ctx, `
		INSERT INTO race_registrations (race_id, user_id, registered_at, status, medical_certificate, is_medical_certificate_approved)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (race_id, user_id) DO UPDATE SET registered_at = $3, status = $4, medical_certificate = $5, is_medical_certificate_approved = $6
		`, race.Id, registration.UserId, registration.RegisteredAt, registration.Status, registration.MedicalCertificate, registration.IsMedicalCertificateApproved)
		if err != nil {
			return kcore.Wrap(err, "error upserting race_registrations table")
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		return kcore.Wrap(err, "error committing transaction")
	}
	return nil
}
