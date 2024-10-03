package main

import (
	"net/http"
	"github.com/sporic/sporic/internal/models"
)

func (app *App) provc_home(w http.ResponseWriter, r *http.Request) {

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

	params := r.URL.Query()
	refno := params.Get("refno")

	application, err := app.applications.FetchByRefNo(refno)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Application = application
	app.render(w, http.StatusOK, "provc_view_application.tmpl", data)
}
