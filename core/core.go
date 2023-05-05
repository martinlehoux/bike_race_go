package core

import (
	"log"

	"github.com/gofrs/uuid"
)

func UUID() uuid.UUID {
	id, err := uuid.NewV4()
	if err != nil {
		log.Fatalf("error generating uuid: %v", err)
	}
	return id
}
