package race

import (
	"bike_race/auth"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/martinlehoux/kagamigo/kauth"
	"github.com/martinlehoux/kagamigo/kcore"
)

var (
	ErrRaceNotFound = errors.New("race not found")
)

type RaceListModel struct {
	Id                    kcore.ID
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
	currentUser, isLoggedIn := auth.UserFromContext(ctx)
	var hasUserRegisteredSelect string
	var queryArgs []interface{}
	if isLoggedIn {
		hasUserRegisteredSelect = `coalesce(bool_or(race_registrations.user_id = $1), false)`
		queryArgs = append(queryArgs, currentUser.Id)
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
	kcore.Expect(err, "error querying races")
	defer rows.Close()

	var races []RaceListModel
	for rows.Next() {
		var hasUserRegistered bool
		var row RaceListModel
		kcore.Expect(rows.Scan(&row.Id, &row.Name, &row.StartAt, &row.IsOpenForRegistration, &row.MaximumParticipants, &row.CoverImage, &row.Organizers, &row.RegisteredCount, &hasUserRegistered), "error scanning races")
		row.CanRegister = isLoggedIn && row.IsOpenForRegistration && row.RegisteredCount < 100 && !hasUserRegistered
		races = append(races, row)
	}
	return races, http.StatusOK, nil
}

type RacePermissionsModel struct {
	CanUpdateDescription    bool
	CanOpenForRegistration  bool
	CanApproveRegistrations bool
}

type RaceDetailModel struct {
	Id                    kcore.ID
	Name                  string
	IsOpenForRegistration bool
	MaximumParticipants   int
	StartAt               time.Time
	CoverImage            string
	Permissions           RacePermissionsModel
}

func RaceDetailQuery(ctx context.Context, conn *pgxpool.Pool, raceId kcore.ID) (RaceDetailModel, int, error) {
	currentUser, _ := auth.UserFromContext(ctx)
	var race RaceDetailModel
	var isCurrentUserOrganizer bool
	err := conn.QueryRow(ctx, `
		SELECT
			races.id, races.name, races.maximum_participants, races.is_open_for_registration, races.start_at, coalesce(races.cover_image_id::text, ''),
			$2::UUID IS NOT NULL AND bool_or(race_organizers.user_id = $2)
		FROM races
		LEFT JOIN race_organizers ON races.id = race_organizers.race_id 
		WHERE races.id = $1
		GROUP BY races.id, races.name
		`, raceId, currentUser.Id).Scan(&race.Id, &race.Name, &race.MaximumParticipants, &race.IsOpenForRegistration, &race.StartAt, &race.CoverImage, &isCurrentUserOrganizer)
	race.Permissions = RacePermissionsModel{
		CanOpenForRegistration:  isCurrentUserOrganizer && race.IsOpenForRegistration,
		CanApproveRegistrations: isCurrentUserOrganizer,
		CanUpdateDescription:    isCurrentUserOrganizer,
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return race, http.StatusNotFound, ErrRaceNotFound
	}
	kcore.Expect(err, "error querying race")

	return race, http.StatusOK, nil
}

type RaceRegistrationPermissionsModel struct {
	CanApprove                   bool
	CanApproveMedicalCertificate bool
}

type RaceRegistrationModel struct {
	User struct {
		Id       kcore.ID
		Username string
	}
	Status                       RaceRegistrationStatus
	RegisteredAt                 time.Time
	MedicalCertificate           *string
	IsMedicalCertificateApproved bool
	Permissions                  RaceRegistrationPermissionsModel
}

func RaceRegistrationsQuery(ctx context.Context, conn *pgxpool.Pool, raceId kcore.ID, racePermissions RacePermissionsModel) ([]RaceRegistrationModel, int, error) {
	var registrations []RaceRegistrationModel
	rows, err := conn.Query(ctx, `
		SELECT
			race_registrations.user_id,
			race_registrations.status,
			race_registrations.registered_at,
			race_registrations.medical_certificate,
			race_registrations.is_medical_certificate_approved,
			users.username
		FROM race_registrations
		LEFT JOIN users ON users.id = race_registrations.user_id
		WHERE race_registrations.race_id = $1
		ORDER BY race_registrations.registered_at ASC
		`, raceId)
	kcore.Expect(err, "error querying race_registrations")
	defer rows.Close()

	for rows.Next() {
		var registration RaceRegistrationModel
		kcore.Expect(rows.Scan(&registration.User.Id, &registration.Status, &registration.RegisteredAt, &registration.MedicalCertificate, &registration.IsMedicalCertificateApproved, &registration.User.Username), "error scanning race_registrations")
		registration.Permissions = RaceRegistrationPermissionsModel{
			CanApprove:                   racePermissions.CanApproveRegistrations && registration.Status == Registered && registration.IsMedicalCertificateApproved,
			CanApproveMedicalCertificate: racePermissions.CanApproveRegistrations && registration.Status == Registered && registration.MedicalCertificate != nil && !registration.IsMedicalCertificateApproved,
		}
		registrations = append(registrations, registration)
	}
	return registrations, http.StatusOK, nil
}

type UserRegistrationModelPermissions struct {
	CanUploadMedicalCertificate bool
}

type UserRegistrationModel struct {
	Status RaceRegistrationStatus
	Race   struct {
		Id   kcore.ID
		Name string
	}
	Permissions UserRegistrationModelPermissions
}

func CurrentUserRegistrationsQuery(ctx context.Context, conn *pgxpool.Pool) ([]UserRegistrationModel, int, error) {
	currentUser, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, http.StatusUnauthorized, kauth.ErrUserNotLoggedIn
	}
	registrations := []UserRegistrationModel{}
	rows, err := conn.Query(ctx, `
		SELECT
			race_registrations.race_id, race_registrations.status, race_registrations.medical_certificate,
			races.name
		FROM
			race_registrations
			INNER JOIN races ON race_registrations.race_id = races.id
		WHERE
			race_registrations.user_id = $1
			`, currentUser.Id)
	kcore.Expect(err, "error querying race_registrations")
	defer rows.Close()
	for rows.Next() {
		var registration UserRegistrationModel
		var medicalCertificate *kcore.File
		kcore.Expect(rows.Scan(&registration.Race.Id, &registration.Status, &medicalCertificate, &registration.Race.Name), "error scanning race_registrations")
		registration.Permissions = UserRegistrationModelPermissions{
			CanUploadMedicalCertificate: registration.Status == Registered && medicalCertificate == nil,
		}
		registrations = append(registrations, registration)
	}
	return registrations, http.StatusOK, nil
}
