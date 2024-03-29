package race

import (
	"bike_race/auth"
	"context"
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/martinlehoux/kagamigo/kauth"
	"github.com/martinlehoux/kagamigo/kcore"
	"github.com/samber/lo"
	"golang.org/x/exp/slog"
)

var (
	ErrUserNotOrganizer = errors.New("user not an organizer")
)

func OrganizeRaceCommand(ctx context.Context, conn *pgxpool.Pool, name string) (int, error) {
	logger := slog.With(slog.String("command", "OrganizeRaceCommand"))
	currentUser, ok := auth.UserFromContext(ctx)
	if !ok {
		logger.Warn(kauth.ErrUserNotLoggedIn.Error())
		return http.StatusUnauthorized, kauth.ErrUserNotLoggedIn
	}
	race, err := NewRace(name)
	if err != nil {
		err = kcore.Wrap(err, "error creating race")
		logger.Warn(err.Error())
		return http.StatusBadRequest, err
	}
	err = race.AddOrganizer(currentUser)
	if err != nil {
		err = kcore.Wrap(err, "error adding organizer")
		logger.Warn(err.Error())
		return http.StatusBadRequest, err
	}
	err = race.Save(ctx, conn)
	if err != nil {
		err = kcore.Wrap(err, "error saving race")
		logger.Error(err.Error())
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}

func OpenRaceForRegistration(ctx context.Context, conn *pgxpool.Pool, raceId kcore.ID, maximumParticipants int) (int, error) {
	logger := slog.With(slog.String("raceId", raceId.String()))
	logger.Info("opening race for registration")
	currentUser, ok := auth.UserFromContext(ctx)
	if !ok {
		slog.Warn(kauth.ErrUserNotLoggedIn.Error())
		return http.StatusUnauthorized, kauth.ErrUserNotLoggedIn
	}
	logger = logger.With(slog.String("userId", currentUser.Id.String()))
	race, err := LoadRace(ctx, conn, raceId)
	if errors.Is(err, pgx.ErrNoRows) {
		logger.Warn(err.Error())
		return http.StatusNotFound, err
	}
	kcore.Expect(err, "")

	if !lo.ContainsBy(race.Organizers, func(userId kcore.ID) bool { return userId == currentUser.Id }) {
		logger.Warn(ErrUserNotOrganizer.Error())
		return http.StatusUnauthorized, ErrUserNotOrganizer
	}

	err = race.OpenForRegistration(maximumParticipants)
	if err != nil {
		err = kcore.Wrap(err, "error opening race for registration")
		logger.Warn(err.Error())
		return http.StatusBadRequest, err
	}

	kcore.Expect(race.Save(ctx, conn), "")

	logger.Info("race opened for registration")
	return http.StatusOK, nil
}

func RegisterForRaceCommand(ctx context.Context, conn *pgxpool.Pool, raceId kcore.ID) (int, error) {
	logger := slog.With(slog.String("command", "RegisterForRaceCommand"), slog.String("raceId", raceId.String()))
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		logger.Warn(kauth.ErrUserNotLoggedIn.Error())
		return http.StatusUnauthorized, kauth.ErrUserNotLoggedIn
	}
	logger = logger.With(slog.String("userId", user.Id.String()))
	race, err := LoadRace(ctx, conn, raceId)
	if errors.Is(err, pgx.ErrNoRows) {
		logger.Warn(err.Error())
		return http.StatusNotFound, err
	}
	kcore.Expect(err, "")

	err = race.Register(user)
	if err != nil {
		err = kcore.Wrap(err, "error registering user")
		logger.Warn(err.Error())
		return http.StatusBadRequest, err
	}

	kcore.Expect(race.Save(ctx, conn), "")

	logger.Info("user registered to race")
	return http.StatusOK, nil
}

func ApproveRaceRegistrationCommand(ctx context.Context, conn *pgxpool.Pool, raceId kcore.ID, userId kcore.ID) (int, error) {
	logger := slog.With(slog.String("command", "ApproveRaceRegistrationCommand"), slog.String("raceId", raceId.String()), slog.String("userId", userId.String()))
	logger.Info("approving user registration")
	currentUser, ok := auth.UserFromContext(ctx)
	if !ok {
		logger.Warn(kauth.ErrUserNotLoggedIn.Error())
		return http.StatusUnauthorized, kauth.ErrUserNotLoggedIn
	}
	race, err := LoadRace(ctx, conn, raceId)
	if errors.Is(err, pgx.ErrNoRows) {
		logger.Warn(err.Error())
		return http.StatusNotFound, err
	}
	kcore.Expect(err, "")

	if !race.IsOrganizer(currentUser) {
		logger.Warn(ErrUserNotOrganizer.Error())
		return http.StatusUnauthorized, ErrUserNotOrganizer
	}
	if !race.IsOpenForRegistration {
		logger.Warn(ErrRegistrationsClosed.Error())
		return http.StatusBadRequest, ErrRegistrationsClosed
	}
	err = race.ApproveRegistration(userId)
	if err != nil {
		err = kcore.Wrap(err, "error approving registration")
		logger.Warn(err.Error())
		return http.StatusBadRequest, err
	}
	kcore.Expect(race.Save(ctx, conn), "")

	logger.Info("user registration approved")
	return http.StatusOK, nil
}

func ApproveRegistrationMedicalCertificateCommand(ctx context.Context, conn *pgxpool.Pool, raceId kcore.ID, userId kcore.ID) (int, error) {
	logger := slog.With(slog.String("command", "ApproveRegistrationMedicalCertificateCommand"), slog.String("raceId", raceId.String()), slog.String("userId", userId.String()))
	logger.Info("approving user registration medical certificate")
	currentUser, ok := auth.UserFromContext(ctx)
	if !ok {
		logger.Warn(kauth.ErrUserNotLoggedIn.Error())
		return http.StatusUnauthorized, kauth.ErrUserNotLoggedIn
	}
	race, err := LoadRace(ctx, conn, raceId)
	if errors.Is(err, pgx.ErrNoRows) {
		logger.Warn(err.Error())
		return http.StatusNotFound, err
	}
	kcore.Expect(err, "")

	if !race.IsOrganizer(currentUser) {
		logger.Warn(ErrUserNotOrganizer.Error())
		return http.StatusUnauthorized, ErrUserNotOrganizer
	}
	err = race.ApproveMedicalCertificate(userId)
	if err != nil {
		err = kcore.Wrap(err, "error approving medical certificate")
		logger.Warn(err.Error())
		return http.StatusBadRequest, err
	}
	kcore.Expect(race.Save(ctx, conn), "")

	logger.Info("user registration medical certificate approved")
	return http.StatusOK, nil
}

func UpdateRaceDescriptionCommand(ctx context.Context, conn *pgxpool.Pool, raceId kcore.ID, clearCoverImage bool, coverImageFile multipart.File) (int, error) {
	logger := slog.With(slog.String("command", "UpdateRaceDescriptionCommand"), slog.String("raceId", raceId.String()))
	logger.Info("updating race description")
	currentUser, ok := auth.UserFromContext(ctx)
	if !ok {
		logger.Warn(kauth.ErrUserNotLoggedIn.Error())
		return http.StatusUnauthorized, kauth.ErrUserNotLoggedIn
	}
	race, err := LoadRace(ctx, conn, raceId)
	if errors.Is(err, pgx.ErrNoRows) {
		logger.Warn(err.Error())
		return http.StatusNotFound, err
	}
	kcore.Expect(err, "")

	if !race.IsOrganizer(currentUser) {
		logger.Warn(ErrUserNotOrganizer.Error())
		return http.StatusUnauthorized, ErrUserNotOrganizer
	}

	if clearCoverImage && race.CoverImage != nil {
		err = race.CoverImage.Delete()
		if err != nil {
			err = kcore.Wrap(err, "error deleting old cover_image")
			logger.Warn(err.Error())
			return http.StatusBadRequest, err
		}
		race.CoverImage = nil
	}

	if coverImageFile != nil {
		coverImage := kcore.NewImage()
		err = coverImage.Save(coverImageFile)
		if err != nil {
			err = kcore.Wrap(err, "error saving cover_image")
			logger.Warn(err.Error())
			return http.StatusBadRequest, err
		}
		race.CoverImage = &coverImage
	}

	kcore.Expect(race.Save(ctx, conn), "")

	logger.Info("updated race description")
	return http.StatusOK, nil
}

func UploadRegistrationMedicalCertificateCommand(ctx context.Context, conn *pgxpool.Pool, raceId kcore.ID, medicalCertificateFile multipart.File, medicalCertificateExt string) (int, error) {
	logger := slog.With(slog.String("command", "UploadRegistrationMedicalCertificateCommand"), slog.String("raceId", raceId.String()))
	logger.Info("uploading registration medical certificate")
	currentUser, ok := auth.UserFromContext(ctx)
	if !ok {
		logger.Warn(kauth.ErrUserNotLoggedIn.Error())
		return http.StatusUnauthorized, kauth.ErrUserNotLoggedIn
	}
	logger = logger.With(slog.String("userId", currentUser.Id.String()))
	race, err := LoadRace(ctx, conn, raceId)
	if errors.Is(err, pgx.ErrNoRows) {
		logger.Warn(err.Error())
		return http.StatusNotFound, err
	}
	kcore.Expect(err, "")

	medicalCertificate := kcore.NewFile(medicalCertificateExt)
	err = medicalCertificate.Save(medicalCertificateFile)
	if err != nil {
		err = kcore.Wrap(err, "error saving medical_certificate")
		logger.Warn(err.Error())
		return http.StatusBadRequest, err
	}
	err = race.UploadMedicalCertificate(currentUser.Id, medicalCertificate)
	if err != nil {
		kcore.Expect(medicalCertificate.Delete(), "error deleting medical certificate")
		err = kcore.Wrap(err, "error uploading medical certificate")
		logger.Warn(err.Error())
		return http.StatusBadRequest, err
	}

	kcore.Expect(race.Save(ctx, conn), "")

	logger.Info("registration medical certificate uploaded")
	return http.StatusOK, nil
}
