package race

import (
	"bike_race/auth"
	"bike_race/core"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/exp/slog"
)

func Router(conn *pgxpool.Pool, baseTpl *template.Template) chi.Router {
	router := chi.NewRouter()
	_, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		err = core.Wrap(err, "error loading location")
		slog.Error(err.Error())
		os.Exit(1)
	}

	router.Post("/organize", organizeRaceRoute(conn))
	router.Post("/{raceId}/upload_medical_certificate", uploadRegistrationMedicalCertificateRoute(conn))
	router.Post("/{raceId}/open_for_registration", openRaceForRegistrationRoute(conn))
	router.Post("/{raceId}/update_description", updateRaceDescriptionRoute(conn))
	router.Post("/{raceId}/register", registerForRaceRoute(conn))
	router.Post("/{raceId}/registrations/{userId}/approve", approveRaceRegistrationRoute(conn))

	router.Get("/registrations", viewCurrentUserRegistrationsRoute(conn, template.Must(template.Must(baseTpl.Clone()).ParseFiles("templates/registrations.html"))))
	router.Get("/{raceId}", viewRaceDetailsRoute(conn, template.Must(template.Must(baseTpl.Clone()).ParseFiles("templates/race.html"))))
	router.Get("/", viewRaceListRoute(conn, template.Must(template.Must(baseTpl.Clone()).ParseFiles("templates/races.html"))))

	return router
}

type RaceTemplateData struct {
	Race              RaceDetailModel
	RaceRegistrations []RaceRegistrationModel
}

func viewRaceDetailsRoute(conn *pgxpool.Pool, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := auth.GetTemplateData(r, RaceTemplateData{})
		ctx := r.Context()
		raceId, err := core.ParseID(chi.URLParam(r, "raceId"))
		if err != nil {
			err = core.Wrap(err, "error parsing raceId")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		raceDetail, code, err := RaceDetailQuery(ctx, conn, raceId)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}
		data.Data.Race = raceDetail
		raceRegistrations, code, err := RaceRegistrationsQuery(ctx, conn, raceId, data.Data.Race.Permissions)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}
		data.Data.RaceRegistrations = raceRegistrations
		core.Expect(tpl.ExecuteTemplate(w, "race.html", data), "error executing template")
	}
}

type RacesTemplateData struct {
	Races []RaceListModel
}

func viewRaceListRoute(conn *pgxpool.Pool, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		data := auth.GetTemplateData(r, RacesTemplateData{})
		races, code, err := RaceListQuery(ctx, conn)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}
		data.Data.Races = races
		core.Expect(tpl.ExecuteTemplate(w, "races.html", data), "error executing template")
	}
}

type CurrentUserRegistrationsTemplateData struct {
	Registrations []UserRegistrationModel
}

func viewCurrentUserRegistrationsRoute(conn *pgxpool.Pool, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		data := auth.GetTemplateData(r, CurrentUserRegistrationsTemplateData{})
		registrations, code, err := CurrentUserRegistrationsQuery(ctx, conn)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}
		data.Data.Registrations = registrations
		core.Expect(tpl.ExecuteTemplate(w, "registrations.html", data), "error executing template")
	}
}

func approveRaceRegistrationRoute(conn *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		raceId, err := core.ParseID(chi.URLParam(r, "raceId"))
		if err != nil {
			err = core.Wrap(err, "error parsing raceId")
			slog.Warn(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userId, err := core.ParseID(chi.URLParam(r, "userId"))
		if err != nil {
			err = core.Wrap(err, "error parsing userId")
			slog.Warn(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		code, err := ApproveRaceRegistrationCommand(ctx, conn, raceId, userId)
		if err != nil {
			http.Error(w, err.Error(), code)
		} else {
			http.Redirect(w, r, "/races/"+raceId.String(), http.StatusSeeOther)
		}
	}
}

func registerForRaceRoute(conn *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		raceId, err := core.ParseID(chi.URLParam(r, "raceId"))
		if err != nil {
			err = core.Wrap(err, "error parsing raceId")
			slog.Warn(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		code, err := RegisterForRaceCommand(ctx, conn, raceId)
		if err != nil {
			http.Error(w, err.Error(), code)
		} else {
			http.Redirect(w, r, "/races/"+raceId.String(), http.StatusSeeOther)
		}
	}
}

func uploadRegistrationMedicalCertificateRoute(conn *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		raceId, err := core.ParseID(chi.URLParam(r, "raceId"))
		if err != nil {
			err = core.Wrap(err, "error parsing raceId")
			slog.Warn(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		medicalCertificateFile, _, err := r.FormFile("medical_certificate")
		if err != nil {
			err = core.Wrap(err, "error parsing medical_certificate")
			slog.Warn(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer medicalCertificateFile.Close()
		code, err := UploadRegistrationMedicalCertificateCommand(ctx, conn, raceId, medicalCertificateFile)
		if err != nil {
			http.Error(w, err.Error(), code)
		} else {
			http.Redirect(w, r, "/races/registrations", http.StatusSeeOther)
		}
	}
}

func updateRaceDescriptionRoute(conn *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		raceId, err := core.ParseID(chi.URLParam(r, "raceId"))
		if err != nil {
			err = core.Wrap(err, "error parsing raceId")
			slog.Warn(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		clearCoverImage := r.FormValue("clear_cover_image")
		coverImageFile, _, err := r.FormFile("cover_image")
		if err != nil && err != http.ErrMissingFile {
			err = core.Wrap(err, "error parsing cover_image")
			slog.Warn(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if coverImageFile != nil {
			defer coverImageFile.Close()
		}
		code, err := UpdateRaceDescriptionCommand(ctx, conn, raceId, coverImageFile != nil || clearCoverImage == "on", coverImageFile)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/races/%s", raceId.String()), http.StatusSeeOther)
	}
}

func openRaceForRegistrationRoute(conn *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func organizeRaceRoute(conn *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		code, err := OrganizeRaceCommand(ctx, conn, r.FormValue("name"))
		if err != nil {
			http.Error(w, err.Error(), code)
		} else {
			http.Redirect(w, r, "/races", http.StatusSeeOther)
		}
	}
}
