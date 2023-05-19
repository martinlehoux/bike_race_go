package core

import (
	"net/http"

	"golang.org/x/exp/slog"
)

func RecoverMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rcv := recover(); rcv != nil {
				if rcv == http.ErrAbortHandler {
					panic(rcv)
				}

				err := rcv.(error)
				slog.Error(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
