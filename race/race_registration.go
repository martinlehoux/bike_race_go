package race

import (
	"bike_race/core"
	"time"
)

type RaceRegistrationStatus string

const (
	Registered RaceRegistrationStatus = "registered"
	Submitted  RaceRegistrationStatus = "submitted"
	Approved   RaceRegistrationStatus = "approved"
)

type RaceRegistration struct {
	UserId                       core.ID
	RegisteredAt                 time.Time
	Status                       RaceRegistrationStatus
	MedicalCertificate           *core.File
	IsMedicalCertificateApproved bool
}

func NewRaceRegistration(userId core.ID) RaceRegistration {
	return RaceRegistration{
		UserId:                       userId,
		Status:                       Registered,
		RegisteredAt:                 time.Now(),
		MedicalCertificate:           nil,
		IsMedicalCertificateApproved: false,
	}
}
