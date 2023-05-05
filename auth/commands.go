package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(ctx context.Context, conn *pgx.Conn, username string, password string) (int, error) {
	user, err := CreateUser(username)
	if err != nil {
		err = fmt.Errorf("error creating user: %w", err)
		log.Println(err)
		return http.StatusBadRequest, err
	}
	err = user.SetPassword("", password)
	if err != nil {
		err = fmt.Errorf("error setting password: %w", err)
		log.Println(err)
		return http.StatusBadRequest, err
	}
	err = user.Save(conn, ctx)
	if err != nil {
		err = fmt.Errorf("error saving user: %w", err)
		log.Println(err)
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}

func Authenticate(ctx context.Context, conn *pgx.Conn, username string, password string) (User, error) {
	var user User
	err := conn.QueryRow(ctx, `
		SELECT id, username, password_hash
		FROM users
		WHERE username = $1
	`, username).Scan(&user.Id, &user.Username, &user.PasswordHash)
	if err == pgx.ErrNoRows {
		err = errors.New("user not found")
		log.Println(err)
		return User{}, err
	} else if err != nil {
		err = fmt.Errorf("error querying user: %w", err)
		log.Println(err)
		return User{}, err
	}
	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		err = errors.New("incorrect password")
		log.Println(err)
		return User{}, err
	} else if err != nil {
		err = fmt.Errorf("error comparing password hash: %w", err)
		log.Println(err)
		return User{}, err
	}
	return user, nil
}
