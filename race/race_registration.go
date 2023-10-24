package race

import (
	"time"

	"github.com/martinlehoux/kagamigo/kcore"
)

type RaceRegistrationStatus string

const (
	Registered RaceRegistrationStatus = "registered"
	Submitted  RaceRegistrationStatus = "submitted"
	Approved   RaceRegistrationStatus = "approved"
)

type RaceRegistration struct {
	UserId                       kcore.ID
	RegisteredAt                 time.Time
	Status                       RaceRegistrationStatus
	MedicalCertificate           *kcore.File
	IsMedicalCertificateApproved bool
}

func NewRaceRegistration(userId kcore.ID) RaceRegistration {
	return RaceRegistration{
		UserId:                       userId,
		Status:                       Registered,
		RegisteredAt:                 time.Now(),
		MedicalCertificate:           nil,
		IsMedicalCertificateApproved: false,
	}
}
