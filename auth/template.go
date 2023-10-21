package auth

import (
	"net/http"

	"github.com/kataras/i18n"
)

type TemplateData[T any] struct {
	Ok          bool
	CurrentUser User
	T           func(format string, args ...any) string
	Data        T
}

type Tr = func(format string, args ...any) string

func GetTr(r *http.Request) Tr {
	lang := "en-GB"
	user, ok := UserFromContext(r.Context())
	if ok {
		lang = user.Language
	}
	return func(format string, args ...any) string { return i18n.Tr(lang, format, args...) }
}

type LoginContext struct {
	LoggedIn bool
	User     User
	Tr       Tr
}

func GetLoginContext(r *http.Request) LoginContext {
	user, ok := UserFromContext(r.Context())
	return LoginContext{
		LoggedIn: ok,
		User:     user,
		Tr:       GetTr(r),
	}
}

func GetTemplateData[T any](r *http.Request, data T) TemplateData[T] {
	user, ok := UserFromContext(r.Context())
	return TemplateData[T]{
		Ok:          ok,
		CurrentUser: user,
		T:           GetTr(r),
		Data:        data,
	}
}
