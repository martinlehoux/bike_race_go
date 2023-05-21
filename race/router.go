package race

import (
	"bike_race/auth"
	"bike_race/core"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
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
	paris, err := time.LoadLocation("Europe/Paris")
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
		loggedInUser, ok := auth.UserFromContext(ctx)
		if !ok {
			auth.Unauthorized(w, errors.New("not authenticated"))
			return
		}
		code, err := OrganizeRaceCommand(ctx, conn, r.FormValue("name"), loggedInUser)
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

	router.Post("/{raceId}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		raceId, err := core.ParseID(chi.URLParam(r, "raceId"))
		if err != nil {
			err = core.Wrap(err, "error parsing raceId")
			slog.Warn(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		race, err := LoadRace(ctx, conn, raceId)
		if err != nil {
			err = core.Wrap(err, "error loading race")
			slog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		loggedInUser, ok := auth.UserFromContext(ctx)
		if !ok {
			auth.Unauthorized(w, errors.New("not authenticated"))
			return
		}
		if org := core.Find(race.Organizers, func(userId core.ID) bool { return userId == loggedInUser.Id }); org == nil {
			auth.Unauthorized(w, errors.New("not an organizer"))
			return
		}
		race.IsOpenForRegistration = r.FormValue("is_open_for_registration") == "on"
		race.StartAt, err = time.ParseInLocation("2006-01-02T15:04", r.FormValue("start_at"), paris)
		if err != nil {
			err = core.Wrap(err, "error parsing start_at")
			slog.Warn(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = race.Save(ctx, conn)
		if err != nil {
			err = core.Wrap(err, "error saving race")
			slog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/races/%s", race.Id.String()), http.StatusSeeOther)
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
