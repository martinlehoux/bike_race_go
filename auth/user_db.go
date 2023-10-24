package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/martinlehoux/kagamigo/kcore"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

func LoadUser(ctx context.Context, conn *pgxpool.Pool, userId kcore.ID) (User, error) {
	var user User
	err := conn.QueryRow(ctx, `
		SELECT id, username, password_hash, language
		FROM users
		WHERE id = $1
	`, userId).Scan(&user.Id, &user.Username, &user.PasswordHash, &user.language)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUserNotFound
	} else if err != nil {
		return User{}, kcore.Wrap(err, "error querying user")
	}
	return user, nil
}

func (user *User) Save(ctx context.Context, conn *pgxpool.Pool) error {
	_, err := conn.Exec(ctx, `
		INSERT INTO users (id, username, password_hash, language)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET username = $2, password_hash = $3, language = $4
	`, user.Id, user.Username, user.PasswordHash, user.Language)
	if err != nil {
		return kcore.Wrap(err, "error inserting user table")
	}
	return nil
}
