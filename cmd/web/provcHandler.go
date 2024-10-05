package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sporic/sporic/internal/models"
)

func (app *App) provc_home(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	if user.IsAnonymous() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if user.Role != models.Provc {
		app.notFound(w)
		return
	}

	var applications []models.Application

	applications, err := app.applications.FetchAll()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Applications = applications
	app.render(w, http.StatusOK, "provc_home.tmpl", data)
}

func (app *App) provc_view_application(w http.ResponseWriter, r *http.Request) {

	user := app.contextGetUser(r)
	if user.IsAnonymous() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if user.Role != models.Provc {
		app.notFound(w)
		return
	}

	params := httprouter.ParamsFromContext(r.Context())
	refno := params.ByName("refno")
	fmt.Printf("refno: %s\n", refno)
	application, err := app.applications.FetchByRefNo(refno)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Application = application
	app.render(w, http.StatusOK, "provc_view_application.tmpl", data)
}
