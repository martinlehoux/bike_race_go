package auth

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/martinlehoux/kagamigo/kcore"
)

type UserListModel struct {
	Username string
}

func UserListQuery(ctx context.Context, conn *pgxpool.Pool) ([]UserListModel, int, error) {
	rows, err := conn.Query(ctx, `SELECT username FROM users`)
	kcore.Expect(err, "error querying users")
	defer rows.Close()

	var users []UserListModel
	for rows.Next() {
		var user UserListModel
		kcore.Expect(rows.Scan(&user.Username), "error scanning users")
		users = append(users, user)
	}

	return users, http.StatusOK, nil
}
