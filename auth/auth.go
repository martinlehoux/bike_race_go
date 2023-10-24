package auth

import (
	"context"

	"github.com/martinlehoux/kagamigo/kauth"
	"github.com/martinlehoux/kagamigo/kcore"
)

type Login = kcore.Login[User]

func LoginFromContext(ctx context.Context) Login {
	return kauth.LoginFromContext[User](ctx)
}

func UserFromContext(ctx context.Context) (User, bool) {
	return kauth.UserFromContext[User](ctx)
}
