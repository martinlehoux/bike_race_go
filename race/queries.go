package race

import (
	"bike_race/auth"
	"bike_race/core"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RaceListModel struct {
	Id                    core.ID
	Name                  string
	StartAt               time.Time
	IsOpenForRegistration bool
	Organizers            string
	RegisteredCount       int
	MaximumParticipants   int
	CoverImage            string
	// Permissions
	CanRegister bool
}

func RaceListQuery(ctx context.Context, conn *pgxpool.Pool) ([]RaceListModel, int, error) {
	loggedInUser, isLoggedIn := auth.UserFromContext(ctx)
	var hasUserRegisteredSelect string
	var queryArgs []interface{}
	if isLoggedIn {
		hasUserRegisteredSelect = `coalesce(bool_or(race_registrations.user_id = $1), false)`
		queryArgs = append(queryArgs, loggedInUser.Id)
	} else {
		hasUserRegisteredSelect = `false`
	}
	rows, err := conn.Query(ctx, fmt.Sprintf(`
		SELECT
			races.id, races.name, races.start_at, races.is_open_for_registration, races.maximum_participants, coalesce(races.cover_image_id::text, ''),
			string_agg(users.username, ', '),
			count(distinct race_registrations.user_id) filter (where race_registrations.user_id is not null),
			%s
		FROM races
		LEFT JOIN race_organizers ON races.id = race_organizers.race_id
		LEFT JOIN users ON race_organizers.user_id = users.id
		LEFT JOIN race_registrations on races.id = race_registrations.race_id
		GROUP BY races.id, races.name
		`, hasUserRegisteredSelect), queryArgs...)
	if err != nil {
		err = core.Wrap(err, "error querying races")
		panic(err)
	}
	defer rows.Close()
	var races []RaceListModel
	for rows.Next() {
		var hasUserRegistered bool
		var row RaceListModel
		err := rows.Scan(&row.Id, &row.Name, &row.StartAt, &row.IsOpenForRegistration, &row.MaximumParticipants, &row.CoverImage, &row.Organizers, &row.RegisteredCount, &hasUserRegistered)
		if err != nil {
			err = core.Wrap(err, "error scanning races")
			panic(err)
		}
		row.CanRegister = isLoggedIn && row.IsOpenForRegistration && row.RegisteredCount < 100 && !hasUserRegistered
		races = append(races, row)
	}
	return races, http.StatusOK, nil
}

type RaceDetailModel struct {
	Id                    core.ID
	Name                  string
	IsOpenForRegistration bool
	MaximumParticipants   int
	StartAt               time.Time
	CoverImage            string
	// Permissions
	CanUpdateDescription   bool
	CanOpenForRegistration bool
	CanAcceptRegistrations bool
}

func RaceDetailQuery(ctx context.Context, conn *pgxpool.Pool, raceId core.ID, loggedInUser auth.User) (RaceDetailModel, int, error) {
	var race RaceDetailModel
	var isLoggedInUserOrganizer bool
	err := conn.QueryRow(ctx, `
		SELECT
			races.id, races.name, races.maximum_participants, races.is_open_for_registration, races.start_at, coalesce(races.cover_image_id::text, ''),
			$2::UUID IS NOT NULL AND bool_or(race_organizers.user_id = $2)
		FROM races
		LEFT JOIN race_organizers ON races.id = race_organizers.race_id 
		WHERE races.id = $1
		GROUP BY races.id, races.name
		`, raceId, loggedInUser.Id).Scan(&race.Id, &race.Name, &race.MaximumParticipants, &race.IsOpenForRegistration, &race.StartAt, &race.CoverImage, &isLoggedInUserOrganizer)
	race.CanOpenForRegistration = isLoggedInUserOrganizer && race.IsOpenForRegistration
	race.CanAcceptRegistrations = isLoggedInUserOrganizer
	race.CanUpdateDescription = isLoggedInUserOrganizer
	if err == pgx.ErrNoRows {
		err = errors.New("race not found")
		return race, http.StatusNotFound, err
	} else if err != nil {
		err = core.Wrap(err, "error querying race")
		panic(err)
	}
	return race, http.StatusOK, nil
}

type RaceRegistrationModel struct {
	UserId       core.ID
	Username     string
	Status       RaceRegistrationStatus
	RegisteredAt time.Time
}

func RaceRegistrationsQuery(ctx context.Context, conn *pgxpool.Pool, raceId core.ID) ([]RaceRegistrationModel, int, error) {
	var registrations []RaceRegistrationModel
	rows, err := conn.Query(ctx, `
		SELECT
			race_registrations.user_id,
			race_registrations.status,
			race_registrations.registered_at,
			users.username
		FROM race_registrations
		LEFT JOIN users ON users.id = race_registrations.user_id
		WHERE race_registrations.race_id = $1
		ORDER BY race_registrations.registered_at ASC
		`, raceId)
	if err != nil {
		err = core.Wrap(err, "error querying race_registrations")
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var registration RaceRegistrationModel
		err := rows.Scan(&registration.UserId, &registration.Status, &registration.RegisteredAt, &registration.Username)
		if err != nil {
			err = core.Wrap(err, "error scanning race_registrations")
			panic(err)
		}
		registrations = append(registrations, registration)
	}
	return registrations, http.StatusOK, nil
}
