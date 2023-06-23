## Development

- go install github.com/cosmtrek/air@latest
- go install github.com/amacneil/dbmate@latest
- go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
- go install honnef.co/go/tools/cmd/staticcheck@latest

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
- image url better id? but hard with pointer
- https://kurtextrem.de/posts/modern-way-of-img
- https://web.dev/image-cdns/

## Logging

- https://betterstack.com/community/guides/logging/logging-in-go/
- Use .WithGroup() to keep additional data in subgroup
- Switch to JSON in prod

## Features

- new way to handle start date (`race.StartAt, err = time.ParseInLocation("2006-01-02T15:04", r.FormValue("start_at"), paris)`)
- redirect to previous url on login
- navbar current url

## Linter

- logging only in middleware, command, query, ...
- do not use ErrNoRows in commands
- log success before return
- no string literal for status
- force use of ParseMultipartForm
- use `http.HandlerFunc` instead of `func(w http.ResponseWriter, r *http.Request)`
- avoid calling function inside core.Expect ?

## Architecture

- race/domain vs race/http? but module name matters...

## TODO

- try to split external styling (width, margin, ...) from internal (row, color, padding, ...)
- pgx recover
- tracing
- linting to remove old patterns
  - http://goast.yuroyoro.net/
  - https://arslan.io/2019/06/13/using-go-analysis-to-write-a-custom-linter/
  - https://arslan.io/2020/07/07/using-go-analysis-to-fix-your-source-code/
- Improve templates
  - https://github.com/valyala/quicktemplate
- define where from slog errors
- HTMX to some point
  - https://htmx.org/docs/
  - https://htmx.org/essays/template-fragments/
  - https://gist.github.com/benpate/f92b77ea9b3a8503541eb4b9eb515d8a
  - how to return errors? for forms?
- experiment with aggregation independance
- try a pattern of function builder to prepare logger in advance etc
- also guards, a bit django like
- UserID vs RaceID vs ...
- .With(middleware.SetHeader("Cache-Control", "max-age=3600")) => not in dev, should it be golang?
- https://github.com/samber/do
- https://github.com/samber/mo

```go
if !ok {
  err := errors.New("user not logged in")
  slog.Warn(err.Error())
  return http.StatusUnauthorized, err
}
```
