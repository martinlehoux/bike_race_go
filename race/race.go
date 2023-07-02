package race

import (
	"bike_race/auth"
	"bike_race/core"
	"errors"
	"time"

	"github.com/samber/lo"
)

var (
	ErrUserNotRegistered                          = errors.New("user is not registered")
	ErrUserAlreadyRegistered                      = errors.New("user is already registered")
	ErrRegistrationWrongStatus                    = errors.New("registration is not in the correct status")
	ErrRegistrationsClosed                        = errors.New("registrations are closed")
	ErrMedicalCertificateNotApproved              = errors.New("medical certificate is not approved")
	ErrMedicalCertificateMissing                  = errors.New("medical certificate is missing")
	ErrMaximumParticipantsMinimumOne              = errors.New("maximum participants must be at least 1")
	ErrMaximumParticipantsLessThanRegisteredUsers = errors.New("maximum participants cannot be less than current number of registered users")
	ErrRaceNameTooShort                           = errors.New("name must be at least 3 characters")
)

type Race struct {
	Id         core.ID
	Name       string
	Organizers []core.ID
	StartAt    time.Time
	// Description
	CoverImage *core.Image
	// Registration
	IsOpenForRegistration bool
	MaximumParticipants   int
	Registrations         map[core.ID]RaceRegistration
}

func NewRace(name string) (Race, error) {
	if len(name) < 3 {
		return Race{}, ErrRaceNameTooShort
	}
	return Race{
		Id:                    core.NewID(),
		Name:                  name,
		Organizers:            []core.ID{},
		IsOpenForRegistration: false,
		Registrations:         map[core.ID]RaceRegistration{},
	}, nil
}

func (race *Race) AddOrganizer(user auth.User) error {
	race.Organizers = append(race.Organizers, user.Id)
	return nil
}

func (race *Race) Register(user auth.User) error {
	_, ok := race.Registrations[user.Id]
	if ok {
		return ErrUserAlreadyRegistered
	}
	if !race.IsOpenForRegistration {
		return ErrRegistrationsClosed
	}
	race.Registrations[user.Id] = NewRaceRegistration(user.Id)
	return nil
}

func (race *Race) OpenForRegistration(maximumParticipants int) error {
	if maximumParticipants <= 0 {
		return ErrMaximumParticipantsMinimumOne
	}
	if len(race.Registrations) > maximumParticipants {
		return ErrMaximumParticipantsLessThanRegisteredUsers
	}
	race.MaximumParticipants = maximumParticipants
	race.IsOpenForRegistration = true
	return nil
}

func (race *Race) ApproveMedicalCertificate(userId core.ID) error {
	registration, ok := race.Registrations[userId]
	if !ok {
		return ErrUserNotRegistered
	}
	if registration.MedicalCertificate == nil {
		return ErrMedicalCertificateMissing
	}
	registration.IsMedicalCertificateApproved = true
	race.Registrations[userId] = registration
	return nil
}

func (race *Race) ApproveRegistration(userId core.ID) error {
	registration, ok := race.Registrations[userId]
	if !ok {
		return ErrUserNotRegistered
	}
	if registration.Status != Registered {
		return ErrRegistrationWrongStatus
	}
	if !registration.IsMedicalCertificateApproved {
		return ErrMedicalCertificateNotApproved
	}
	registration.Status = Approved
	race.Registrations[userId] = registration
	return nil
}

func (race *Race) UploadMedicalCertificate(userId core.ID, medicalCertificate core.File) error {
	registration, ok := race.Registrations[userId]
	if !ok {
		return ErrUserNotRegistered
	}
	if registration.Status != Registered {
		return ErrRegistrationWrongStatus
	}

	if registration.MedicalCertificate != nil {
		err := registration.MedicalCertificate.Delete()
		if err != nil {
			return core.Wrap(err, "error deleting old medical certificate")
		}
	}

	registration.MedicalCertificate = &medicalCertificate
	registration.IsMedicalCertificateApproved = false
	race.Registrations[userId] = registration
	return nil
}

func (race Race) IsOrganizer(user auth.User) bool {
	return lo.ContainsBy(race.Organizers, func(id core.ID) bool { return id == user.Id })
}
