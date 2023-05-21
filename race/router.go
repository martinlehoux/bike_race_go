package race

import (
	"bike_race/auth"
	"bike_race/core"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/exp/slog"
)

type RacesTemplateData struct {
	LoggedInUser auth.User
	Races        []RaceListModel
}

type RaceTemplateData struct {
	LoggedInUser auth.User
	Race         RaceDetailModel
}

func Router(conn *pgx.Conn, baseTpl *template.Template) chi.Router {
	router := chi.NewRouter()
	_, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		err = core.Wrap(err, "error loading location")
		slog.Error(err.Error())
		os.Exit(1)
	}

	raceListTpl := template.Must(template.Must(baseTpl.Clone()).ParseFiles("templates/races.html"))
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		loggedInUser, _ := auth.UserFromContext(ctx)
		races, code, err := RaceListQuery(ctx, conn)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}
		templateData := RacesTemplateData{
			LoggedInUser: loggedInUser,
			Races:        races,
		}
		err = raceListTpl.ExecuteTemplate(w, "races.html", templateData)
		if err != nil {
			err = core.Wrap(err, "error executing template")
			panic(err)
		}
	})

	router.Post("/organize", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		code, err := OrganizeRaceCommand(ctx, conn, r.FormValue("name"))
		if err != nil {
			http.Error(w, err.Error(), code)
		} else {
			http.Redirect(w, r, "/races", http.StatusSeeOther)
		}
	})

	raceDetailTpl := template.Must(template.Must(baseTpl.Clone()).ParseFiles("templates/race.html"))
	router.Get("/{raceId}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		raceId, err := core.ParseID(chi.URLParam(r, "raceId"))
		if err != nil {
			err = core.Wrap(err, "error parsing raceId")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		loggedInUser, _ := auth.UserFromContext(ctx)
		templateData := RaceTemplateData{
			LoggedInUser: loggedInUser,
		}
		raceDetail, code, err := RaceDetailQuery(ctx, conn, raceId, loggedInUser)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}
		templateData.Race = raceDetail
		err = raceDetailTpl.ExecuteTemplate(w, "race.html", templateData)
		if err != nil {
			err = core.Wrap(err, "error executing template")
			panic(err)
		}
	})

	router.Post("/{raceId}/open_for_registration", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		raceId, err := core.ParseID(chi.URLParam(r, "raceId"))
		if err != nil {
			err = core.Wrap(err, "error parsing raceId")
			slog.Warn(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		maximumParticipants, err := strconv.Atoi(r.FormValue("maximum_participants"))
		if err != nil {
			err = core.Wrap(err, "error parsing maximum_participants")
			slog.Warn(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		code, err := OpenRaceForRegistration(ctx, conn, raceId, maximumParticipants)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		} else {
			http.Redirect(w, r, fmt.Sprintf("/races/%s", raceId.String()), http.StatusSeeOther)
		}
	})

	router.Post("/{raceId}/register", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, ok := auth.UserFromContext(ctx)
		if !ok {
			auth.Unauthorized(w, errors.New("not authenticated"))
			return
		}

		raceId, err := core.ParseID(chi.URLParam(r, "raceId"))
		if err != nil {
			err = core.Wrap(err, "error parsing raceId")
			slog.Warn(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		code, err := RegisterForRaceCommand(ctx, conn, raceId, user)
		if err != nil {
			http.Error(w, err.Error(), code)
		} else {
			http.Redirect(w, r, "/races/"+raceId.String(), http.StatusSeeOther)
		}
	})

	return router
}
