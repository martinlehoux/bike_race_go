package race

import (
	"bike_race/auth"
	"bike_race/core"
	"errors"
	"time"
)

type Race struct {
	Id         core.ID
	Name       string
	Organizers []core.ID
	StartAt    time.Time
	// Registration
	IsOpenForRegistration bool
	MaximumParticipants   int
	Registrations         map[core.ID]RaceRegistration
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
		return errors.New("user already registered")
	}
	if !race.IsOpenForRegistration {
		return errors.New("registration is closed")
	}
	race.Registrations[user.Id] = NewRaceRegistration(user.Id)
	return nil
}

func (race *Race) OpenForRegistration(maximumParticipants int) error {
	if maximumParticipants <= 0 {
		return errors.New("maximum participants must be at least 1")
	}
	if len(race.Registrations) > maximumParticipants {
		return errors.New("maximum participants cannot be less than current number of registered users")
	}
	race.MaximumParticipants = maximumParticipants
	race.IsOpenForRegistration = true
	return nil
}

func (race *Race) ApproveRegistration(userId core.ID) error {
	registration, ok := race.Registrations[userId]
	if !ok {
		return errors.New("user is not registered")
	}
	if registration.Status != Registered {
		return errors.New("registration can't be approved")
	}
	registration.Status = Approved
	race.Registrations[userId] = registration
	return nil
}

func (race Race) CanAcceptRegistration(user auth.User) bool {
	return race.IsOpenForRegistration && core.Find(race.Organizers, func(id core.ID) bool { return id == user.Id }) != nil
}
