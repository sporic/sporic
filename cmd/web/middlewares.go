package main

import (
	"net/http"

	"github.com/sporic/sporic/internal/models"
)

func (app *App) authenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.sessionManager.Exists(r.Context(), "authenticatedUserID") {
			r = app.contextSetUser(r, models.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		id, ok := app.sessionManager.Get(r.Context(), "authenticatedUserID").(int)
		if !ok {
			r = app.contextSetUser(r, models.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}
		user, err := app.users.Get(id)
		if err == models.ErrRecordNotFound {
			r = app.contextSetUser(r, models.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			app.serverError(w, err)
		}

		r = app.contextSetUser(r, user)
		next.ServeHTTP(w, r)

	})
}
