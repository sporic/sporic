package main

import (
	"net/http"

	"github.com/sporic/sporic/internal/models"
)

func (app *App) authenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.sessionManager.Exists(r.Context(), "authenticatedUserID") {
			app.errorLog.Println("No authenticatedUserID found in session")
			r = app.contextSetUser(r, models.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		id, ok := app.sessionManager.Get(r.Context(), "authenticatedUserID").(int)
		if !ok {
			app.errorLog.Println("authenticatedUserID is not an int")
			r = app.contextSetUser(r, models.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}
		user, err := app.users.Get(id)
		if err == models.ErrRecordNotFound {
			app.errorLog.Println("authenticatedUserID not found in database")
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
