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

func Data[T any](r *http.Request, data T) TemplateData[T] {
	user, ok := UserFromContext(r.Context())
	lang := "en"
	return TemplateData[T]{
		Ok:          ok,
		CurrentUser: user,
		T:           func(format string, args ...any) string { return i18n.Tr(lang, format, args...) },
		Data:        data,
	}
}
