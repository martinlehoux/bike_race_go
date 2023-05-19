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

func Router(conn *pgx.Conn, tpl *template.Template) chi.Router {
	router := chi.NewRouter()
	paris, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		err = core.Wrap(err, "error loading location")
		slog.Error(err.Error())
		os.Exit(1)
	}

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
		err = tpl.ExecuteTemplate(w, "races.html", templateData)
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
		err = tpl.ExecuteTemplate(w, "race.html", templateData)
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
