package main

import (
	"net/http"

	"github.com/sporic/sporic/internal/models"
	"github.com/sporic/sporic/internal/validator"

	_ "github.com/go-sql-driver/mysql"
)

type loginForm struct {
	Username            string `form:"username"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *App) login(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = loginForm{}
	app.render(w, http.StatusOK, "login.tmpl", data)
}

func (app *App) loginPost(w http.ResponseWriter, r *http.Request) {
	var form loginForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusUnprocessableEntity)
		return
	}

	form.CheckField(validator.NotBlank(form.Username), "username", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
		return
	}

	id, err := app.users.Authenticate(form.Username, form.Password)
	if err != models.ErrInvalidCredentials {
		form.AddNonFieldError("Invalid username/password")
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnauthorized, "login.tmpl", data)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
    if err != nil {
        app.serverError(w, err)
        return
    }
    

    app.sessionManager.Put(r.Context(), "authenticatedUserID", id)

	data := app.newTemplateData(r)
	data.Form = form
	app.render(w, http.StatusOK, "login.tmpl", data)
}
