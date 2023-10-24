package main

import (
	"bike_race/auth"
	"bike_race/config"
	"bike_race/race"
	"context"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kataras/i18n"
	"github.com/martinlehoux/kagamigo/kauth"
	"github.com/martinlehoux/kagamigo/kcore"
	"github.com/riandyrn/otelchi"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"golang.org/x/exp/slog"
)

func getTracerProvider(ctx context.Context, serviceName string) *trace.TracerProvider {
	// version, env, ...
	providerResource, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(semconv.ServiceName(serviceName)),
	)
	kcore.Expect(err, "error creating resource")

	httpClient := otlptracehttp.NewClient(otlptracehttp.WithEndpoint(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")), otlptracehttp.WithInsecure())
	httpExporter, err := otlptrace.New(ctx, httpClient)
	kcore.Expect(err, "error creating exporter")

	return trace.NewTracerProvider(
		trace.WithBatcher(httpExporter,
			trace.WithMaxQueueSize(1), // Dev only
		),
		trace.WithResource(providerResource),
	)
}

func main() {
	i18n.SetDefaultLanguage("en-US")
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	ctx := context.Background()
	conf := config.LoadConfig()
	conn := config.LoadDatabasePool(ctx, conf)
	defer conn.Close()

	serviceName := "bike_race"
	tracerProvider := getTracerProvider(ctx, serviceName)
	otel.SetTracerProvider(tracerProvider)
	defer tracerProvider.Shutdown(ctx) //nolint:errcheck

	router := chi.NewRouter()
	router.Use(kcore.RecoverMiddleware)
	router.Use(otelchi.Middleware(serviceName)) // otelchi.WithChiRoutes(router)
	loadUser := func(ctx context.Context, userId kcore.ID) (any, error) {
		user, err := auth.LoadUser(ctx, conn, userId)
		return user, kcore.Wrap(err, "error loading user")
	}
	router.Use(kauth.CookieAuthMiddleware(loadUser, conf.Auth))

	router.With(middleware.SetHeader("Cache-Control", "max-age=3600")).Handle("/favicon.ico", http.FileServer(http.Dir("static")))
	router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	router.Handle("/media/*", http.StripPrefix("/media/", http.FileServer(http.Dir("media"))))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		login := auth.LoginFromContext(ctx)
		page := IndexPage(login)
		kcore.RenderPage(r.Context(), page, w)
	})

	router.Mount("/users", auth.Router(conn, conf))
	router.Mount("/races", race.Router(conn))

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		login := auth.LoginFromContext(ctx)
		page := NotFoundPage(login)
		kcore.RenderPage(r.Context(), page, w)
	})

	slog.Info("listening on http://localhost:3000")
	server := http.Server{
		Addr:              ":3000",
		WriteTimeout:      1 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		Handler:           router,
	}
	err := server.ListenAndServe()
	kcore.Expect(err, "error listening and serving")
}
