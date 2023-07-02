package main

import (
	"bike_race/auth"
	"bike_race/config"
	"bike_race/core"
	"bike_race/race"
	"context"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kataras/i18n"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"golang.org/x/exp/slog"
)

func tracerMiddleware(next http.Handler) http.Handler {
	tracer := otel.Tracer("tracer")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, span := tracer.Start(ctx, r.URL.Path)
		defer span.End()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	i18n.SetDefaultLanguage("en-US")
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	ctx := context.Background()
	conf := config.LoadConfig()
	conn := config.LoadDatabasePool(ctx, conf)
	defer conn.Close()
	baseTpl := template.Must(template.New("").ParseGlob("templates/base/*.html"))

	client := otlptracehttp.NewClient(otlptracehttp.WithEndpoint(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")), otlptracehttp.WithInsecure())
	exporter, err := otlptrace.New(ctx, client)
	core.Expect(err, "error creating exporter")
	tracerProvider := trace.NewTracerProvider(trace.WithBatcher(exporter), trace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName("bike_race"))))
	otel.SetTracerProvider(tracerProvider)
	defer tracerProvider.Shutdown(ctx) //nolint:errcheck

	router := chi.NewRouter()
	router.Use(core.RecoverMiddleware)
	router.Use(tracerMiddleware)
	router.Use(auth.CookieAuthMiddleware(conn, conf))

	router.With(middleware.SetHeader("Cache-Control", "max-age=3600")).Handle("/favicon.ico", http.FileServer(http.Dir("static")))
	router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	router.Handle("/media/*", http.StripPrefix("/media/", http.FileServer(http.Dir("media"))))

	indexTpl := template.Must(template.Must(baseTpl.Clone()).ParseFiles("templates/index.html"))
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data := auth.GetTemplateData(r, struct{}{})
		core.ExecuteTemplate(w, *indexTpl, "index.html", data)
	})

	router.Mount("/users", auth.Router(conn, baseTpl, conf))
	router.Mount("/races", race.Router(conn, baseTpl))

	tpl404 := template.Must(template.Must(baseTpl.Clone()).ParseFiles("templates/404.html"))
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		data := auth.GetTemplateData(r, struct{}{})
		core.Expect(tpl404.ExecuteTemplate(w, "404.html", data), "error executing template")
	})

	slog.Info("listening on http://localhost:3000")
	server := http.Server{
		Addr:              ":3000",
		WriteTimeout:      1 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		Handler:           router,
	}
	err = server.ListenAndServe()
	core.Expect(err, "error listening and serving")
}
