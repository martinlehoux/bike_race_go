package race

import (
	"bike_race/auth"
	"bike_race/core"
	"errors"
	"fmt"
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
	Races []RacesTemplateDataRow
}

func Router(conn *pgx.Conn, tpl *template.Template) chi.Router {
	router := chi.NewRouter()

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, ok := ctx.Value("user").(auth.User)
		if !ok {
			auth.Unauthorized(w, errors.New("not authenticated"))
			return
		}
		templateData := RacesTemplateData{}
		rows, err := conn.Query(ctx, `
		SELECT races.id, races.name, string_agg(users.username, ', ')
		FROM races
		LEFT JOIN race_organizers ON races.id = race_organizers.race_id
		LEFT JOIN users ON race_organizers.user_id = users.id
		GROUP BY races.id, races.name
		`)
		if err != nil {
			err = fmt.Errorf("error querying races: %w", err)
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var row RacesTemplateDataRow
			err := rows.Scan(&row.RaceId, &row.RaceName, &row.Organizers)
			if err != nil {
				err = fmt.Errorf("error scanning races: %w", err)
				log.Fatal(err)
			}
			templateData.Races = append(templateData.Races, row)
		}
		err = tpl.ExecuteTemplate(w, "races.html", templateData)
		if err != nil {
			err = fmt.Errorf("error executing template: %w", err)
			log.Fatal(err)
		}
	})

	router.Post("/organize", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, ok := ctx.Value("user").(auth.User)
		if !ok {
			auth.Unauthorized(w, errors.New("not authenticated"))
			return
		}
		code, err := OrganizeRace(ctx, conn, r.FormValue("name"), user)
		if err != nil {
			w.WriteHeader(code)
			w.Write([]byte(err.Error()))
		} else {
			http.Redirect(w, r, "/races", http.StatusSeeOther)
		}
	})

	return router
}
