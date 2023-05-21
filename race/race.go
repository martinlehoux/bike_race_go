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

func (race *Race) OpenForRegistration(maximumParticipants int) error {
	if maximumParticipants <= 0 {
		return errors.New("maximum participants must be at least 1")
	}
	if len(race.RegisteredUsers) > maximumParticipants {
		return errors.New("maximum participants cannot be less than current number of registered users")
	}
	race.MaximumParticipants = maximumParticipants
	race.IsOpenForRegistration = true
	return nil
}
