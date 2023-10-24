package auth

import (
	"net/http"

	"github.com/martinlehoux/kagamigo/kauth"
	"github.com/martinlehoux/kagamigo/kcore"
)

type TemplateData[T any] struct {
	Ok          bool
	CurrentUser User
	T           func(format string, args ...any) string
	Data        T
}

func GetTemplateData[T any](r *http.Request, data T) TemplateData[T] {
	user, ok := kauth.UserFromContext[User](r.Context())
	return TemplateData[T]{
		Ok:          ok,
		CurrentUser: user,
		T:           kcore.GetTr(user),
		Data:        data,
	}
}
