## Development

- go install github.com/cosmtrek/air@latest
- go install github.com/amacneil/dbmate@latest
- go install github.com/fzipp/gocyclo/cmd/gocyclo@latest

```
dbmate up
air
cd webapp && npx tailwindcss -i ../static/base.css -o ../static/index.css --watch
```

## Secrets

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable
DBMATE_MIGRATIONS_DIR=migrations/
DBMATE_SCHEMA_FILE=schema.sql
COOKIE_SECRET=`head -c32 </dev/urandom | xxd -p -u`
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318
```

## Images

- https://pkg.go.dev/image
- https://go.dev/blog/image
- https://go.dev/blog/image-draw

## Logging

- https://betterstack.com/community/guides/logging/logging-in-go/

## TODO

- , HTTP 500 (middleware?), setting to display or not http 500, recover
- pgx recover
- fix trailing /
- export queries
- error page / error message
- CSRF
- 404 with NotFound, 400 with message ?
- let's start styling
  - https://github.com/da-revo/go-templating-with-tailwindcss
- handle time zones
- tracing
- linting to remove old patterns
  - not log.Fatal
  - not log.\*
  - https://life.wongnai.com/writing-a-custom-go-vet-for-better-code-standard-7dc8296b5513
- i18n & templates
  - https://github.com/kataras/i18n
  - https://github.com/vorlif/spreak
- go vet ./... (CI)
