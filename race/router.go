package race

import (
	"bike_race/auth"
	"bike_race/core"
	"errors"
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type RacesTemplateDataRow struct {
	RaceId     core.ID
	RaceName   string
	Organizers string
}

type RacesTemplateData struct {
	LoggedInUser auth.User
	Races        []RacesTemplateDataRow
}

func Router(conn *pgx.Conn, tpl *template.Template) chi.Router {
	router := chi.NewRouter()

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		loggedInUser, _ := auth.UserFromContext(ctx)
		templateData := RacesTemplateData{
			LoggedInUser: loggedInUser,
		}
		rows, err := conn.Query(ctx, `
		SELECT races.id, races.name, string_agg(users.username, ', ')
		FROM races
		LEFT JOIN race_organizers ON races.id = race_organizers.race_id
		LEFT JOIN users ON race_organizers.user_id = users.id
		GROUP BY races.id, races.name
		`)
		if err != nil {
			err = core.Wrap(err, "error querying races")
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var row RacesTemplateDataRow
			err := rows.Scan(&row.RaceId, &row.RaceName, &row.Organizers)
			if err != nil {
				err = core.Wrap(err, "error scanning races")
				log.Fatal(err)
			}
			templateData.Races = append(templateData.Races, row)
		}
		err = tpl.ExecuteTemplate(w, "races.html", templateData)
		if err != nil {
			err = core.Wrap(err, "error executing template")
			log.Fatal(err)
		}
	})

	router.Post("/organize", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		loggedInUser, ok := auth.UserFromContext(ctx)
		if !ok {
			auth.Unauthorized(w, errors.New("not authenticated"))
			return
		}
		code, err := OrganizeRace(ctx, conn, r.FormValue("name"), loggedInUser)
		if err != nil {
			w.WriteHeader(code)
			w.Write([]byte(err.Error()))
		} else {
			http.Redirect(w, r, "/races", http.StatusSeeOther)
		}
	})

	router.Get("/{raceId}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		raceId, err := core.ParseID(chi.URLParam(r, "raceId"))
		if err != nil {
			err = core.Wrap(err, "error parsing raceId")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		var raceName string
		err = conn.QueryRow(ctx, `
		SELECT races.name FROM races WHERE races.id = $1
		`, raceId).Scan(&raceName)
		if err != nil {
			err = core.Wrap(err, "error querying race")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		templateData := struct{ RaceName string }{RaceName: raceName}
		err = tpl.ExecuteTemplate(w, "race.html", templateData)
		if err != nil {
			err = core.Wrap(err, "error executing template")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	})

	return router
}
