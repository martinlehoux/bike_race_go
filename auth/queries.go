package auth

import (
	"bike_race/core"
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserListModel struct {
	Username string
}

func UserListQuery(ctx context.Context, conn *pgxpool.Pool) ([]UserListModel, int, error) {
	rows, err := conn.Query(ctx, `SELECT username FROM users`)
	if err != nil {
		err = core.Wrap(err, "error querying users")
		panic(err)
	}
	defer rows.Close()

	var users []UserListModel
	for rows.Next() {
		var user UserListModel
		err := rows.Scan(&user.Username)
		if err != nil {
			err = core.Wrap(err, "error scanning users")
			panic(err)
		}
		users = append(users, user)
	}

	return users, http.StatusOK, nil
}
