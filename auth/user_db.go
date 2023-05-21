package auth

import (
	"bike_race/core"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func LoadUser(ctx context.Context, conn *pgx.Conn, userId core.ID) (User, error) {
	var user User
	err := conn.QueryRow(ctx, `
		SELECT id, username, password_hash
		FROM users
		WHERE id = $1
	`, userId).Scan(&user.Id, &user.Username, &user.PasswordHash)
	if err == pgx.ErrNoRows {
		return User{}, errors.New("user not found")
	} else if err != nil {
		return User{}, core.Wrap(err, "error querying user")
	}
	return user, nil
}

func (user *User) Save(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, `
		INSERT INTO users (id, username, password_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET username = $2, password_hash = $3
	`, user.Id, user.Username, user.PasswordHash)
	return err
}