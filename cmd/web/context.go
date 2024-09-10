package main

import (
	"context"
	"net/http"

	"github.com/sporic/sporic/internal/models"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *App) contextSetUser(r *http.Request, user *models.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *App) contextGetUser(r *http.Request) *models.User {
	user, ok := r.Context().Value(userContextKey).(*models.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}

func (app *App) authenticateUser(r *http.Request) (bool, error) {
	_, ok := r.Context().Value(userContextKey).(*models.User)
	if ok {
		return true, nil
	}

	if !app.sessionManager.Exists(r.Context(), "authenticatedUserID") {
		return false, nil
	}

	val, ok := app.sessionManager.Get(r.Context(), "authenticatedUserID").(int)
	if !ok {
		return false, nil
	}

	user, err := app.users.Get(val)
	if err == models.ErrRecordNotFound {
		app.sessionManager.Remove(r.Context(), "authenticatedUserID")
		return false, nil
	} else if err != nil {
		return false, err
	}

	app.contextSetUser(r, user)
	return true, nil
}
