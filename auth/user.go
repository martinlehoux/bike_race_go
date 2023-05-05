package auth

import (
	"bike_race/core"
	"context"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id           uuid.UUID
	Username     string
	PasswordHash []byte
}

func CreateUser(username string) (User, error) {
	var user User
	if len(username) < 3 {
		return user, errors.New("username must be at least 3 characters")
	}
	user.Id = core.UUID()
	user.Username = username
	return user, nil
}

func (user *User) Save(conn *pgx.Conn, ctx context.Context) error {
	_, err := conn.Exec(ctx, `
		INSERT INTO users (id, username, password_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET username = $2, password_hash = $3
	`, user.Id, user.Username, user.PasswordHash)
	return err
}

func (user *User) SetPassword(oldPassword string, newPassword string) error {
	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(oldPassword)) != nil {
		return errors.New("incorrect password")
	}
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = newPasswordHash
	return nil
}
