package core

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/gofrs/uuid"
)

type ID struct {
	uuid.UUID
}

func NewID() ID {
	id, err := uuid.NewV4()
	if err != nil {
		log.Fatalf("error generating uuid: %v", err)
	}
	return ID{id}
}

func (id ID) String() string {
	return base64.URLEncoding.EncodeToString(id.Bytes())
}

func ParseID(value string) (ID, error) {
	bytes, err := base64.URLEncoding.DecodeString(value)
	if err != nil {
		return ID{}, fmt.Errorf("error decoding value: %w", err)
	}
	id, err := uuid.FromBytes(bytes)
	if err != nil {
		return ID{}, fmt.Errorf("error parsing uuid: %w", err)
	}
	return ID{id}, nil
}
