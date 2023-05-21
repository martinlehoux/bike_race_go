package race

import (
	"bike_race/auth"
	"bike_race/core"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

type RaceListModel struct {
	Id                    core.ID
	Name                  string
	StartAt               time.Time
	IsOpenForRegistration bool
	Organizers            string
	RegisteredCount       int
}

func RaceListQuery(ctx context.Context, conn *pgx.Conn) ([]RaceListModel, int, error) {
	rows, err := conn.Query(ctx, `
		SELECT
			races.id, races.name, races.start_at, races.is_open_for_registration,
			string_agg(users.username, ', '),
			count(distinct race_registered_users.user_id) filter (where race_registered_users.user_id is not null)
		FROM races
		LEFT JOIN race_organizers ON races.id = race_organizers.race_id
		LEFT JOIN users ON race_organizers.user_id = users.id
		LEFT JOIN race_registered_users on races.id = race_registered_users.race_id
		GROUP BY races.id, races.name
		`)
	if err != nil {
		err = core.Wrap(err, "error querying races")
		panic(err)
	}
	defer rows.Close()
	var races []RaceListModel
	for rows.Next() {
		var row RaceListModel
		err := rows.Scan(&row.Id, &row.Name, &row.StartAt, &row.IsOpenForRegistration, &row.Organizers, &row.RegisteredCount)
		if err != nil {
			err = core.Wrap(err, "error scanning races")
			panic(err)
		}
		races = append(races, row)
	}
	return races, http.StatusOK, nil
}

type RaceDetailModel struct {
	Id                    core.ID
	Name                  string
	IsOpenForRegistration bool
	StartAt               time.Time
	IsEditable            bool
}

func RaceDetailQuery(ctx context.Context, conn *pgx.Conn, raceId core.ID, loggedInUser auth.User) (RaceDetailModel, int, error) {
	var race RaceDetailModel
	err := conn.QueryRow(ctx, `
		SELECT races.id, races.name, $2::UUID IS NOT NULL AND bool_or(race_organizers.user_id = $2) AS is_editable, races.is_open_for_registration, races.start_at
		FROM races
		LEFT JOIN race_organizers ON races.id = race_organizers.race_id 
		WHERE races.id = $1
		GROUP BY races.id, races.name
		`, raceId, loggedInUser.Id).Scan(&race.Id, &race.Name, &race.IsEditable, &race.IsOpenForRegistration, &race.StartAt)
	if err == pgx.ErrNoRows {
		err = errors.New("race not found")
		return race, http.StatusNotFound, err
	} else if err != nil {
		err = core.Wrap(err, "error querying race")
		panic(err)
	}
	return race, http.StatusOK, nil
}
