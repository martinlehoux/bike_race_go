package race

import (
	"bike_race/auth"
	"bike_race/core"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type RacesTemplateDataRow struct {
	RaceId     core.ID
	RaceName   string
	StartAt    time.Time
	Organizers string
}

type RacesTemplateData struct {
	LoggedInUser auth.User
	Races        []RacesTemplateDataRow
}

type RaceTemplateData struct {
	LoggedInUser          auth.User
	RaceId                core.ID
	Name                  string
	IsOpenForRegistration bool
	StartAt               time.Time
	IsEditable            bool
}

func Router(conn *pgx.Conn, tpl *template.Template) chi.Router {
	router := chi.NewRouter()
	paris, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		err = core.Wrap(err, "error loading location")
		log.Fatal(err)
	}

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		loggedInUser, _ := auth.UserFromContext(ctx)
		templateData := RacesTemplateData{
			LoggedInUser: loggedInUser,
		}
		rows, err := conn.Query(ctx, `
		SELECT races.id, races.name, races.start_at, string_agg(users.username, ', ')
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
			err := rows.Scan(&row.RaceId, &row.RaceName, &row.StartAt, &row.Organizers)
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
		loggedInUser, _ := auth.UserFromContext(ctx)
		templateData := RaceTemplateData{
			LoggedInUser: loggedInUser,
		}
		err = conn.QueryRow(ctx, `
		SELECT races.id, races.name, $2::UUID IS NOT NULL AND bool_or(race_organizers.user_id = $2) AS is_editable, races.is_open_for_registration, races.start_at
		FROM races
		LEFT JOIN race_organizers ON races.id = race_organizers.race_id 
		WHERE races.id = $1
		GROUP BY races.id, races.name
		`, raceId, loggedInUser.Id).Scan(&templateData.RaceId, &templateData.Name, &templateData.IsEditable, &templateData.IsOpenForRegistration, &templateData.StartAt)
		if err != nil {
			err = core.Wrap(err, "error querying race")
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		err = tpl.ExecuteTemplate(w, "race.html", templateData)
		if err != nil {
			err = core.Wrap(err, "error executing template")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	})

	router.Post("/{raceId}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		raceId, err := core.ParseID(chi.URLParam(r, "raceId"))
		if err != nil {
			err = core.Wrap(err, "error parsing raceId")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		race, err := LoadRace(conn, ctx, raceId)
		if err != nil {
			err = core.Wrap(err, "error loading race")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		loggedInUser, ok := auth.UserFromContext(ctx)
		if !ok {
			auth.Unauthorized(w, errors.New("not authenticated"))
			return
		}
		if org := core.Find(race.Organizers, func(user auth.User) bool { return user.Id == loggedInUser.Id }); org == nil {
			auth.Unauthorized(w, errors.New("not an organizer"))
			return
		}
		race.IsOpenForRegistration = r.FormValue("is_open_for_registration") == "on"
		race.StartAt, err = time.ParseInLocation("2006-01-02T15:04", r.FormValue("start_at"), paris)
		if err != nil {
			err = core.Wrap(err, "error parsing start_at")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		err = race.Save(conn, ctx)
		if err != nil {
			err = core.Wrap(err, "error save race")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/races/%s", race.Id.String()), http.StatusSeeOther)
	})

	return router
}
