## Development

- go install github.com/cosmtrek/air@latest
- go install github.com/amacneil/dbmate@latest

```
dbmate up
air
```

## Images

- https://pkg.go.dev/image
- https://go.dev/blog/image
- https://go.dev/blog/image-draw

## TODO

- not log.Fatal, HTTP 500 (middleware?), setting to display or not http 500, recover
- pgx recover
- fix trailing /
- export queries
- error page / error message
- CSRF
- structured logger, better log middleware, https://github.com/go-chi/httplog
- 404 with NotFound
- let's start styling
- handle time zones
