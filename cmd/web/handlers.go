package main

import (
	"net/http"
	"strconv"

	"github.com/sporic/sporic/internal/models"
	"github.com/sporic/sporic/internal/validator"
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
	if err == models.ErrInvalidCredentials {
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
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func (app *App) logout(w http.ResponseWriter, r *http.Request) {

	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *App) home(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	if user.IsAnonymous() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if user.Role == models.AdminUser {
		http.Redirect(w, r, "/admin_home", http.StatusSeeOther)
		return
	}
	if user.Role == models.FacultyUser {
		http.Redirect(w, r, "/faculty_home", http.StatusSeeOther)
		return
	}
	app.notFound(w)
}

func (app *App) admin_home(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	if user.IsAnonymous() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if user.Role != models.AdminUser {
		app.notFound(w)
		return
	}
	data := app.newTemplateData(r)
	app.render(w, http.StatusOK, "admin_home.tmpl", data)
}

func (app *App) faculty_home(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	if user.IsAnonymous() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if user.Role != models.FacultyUser {
		app.notFound(w)
		return
	}
	data := app.newTemplateData(r)
	app.render(w, http.StatusOK, "faculty_home.tmpl", data)
}

type newProjectForm struct {
	ActivityType         string `form:"activity_type"`
	FinancialYear        int    `form:"financial_year"`
	EstimatedAmt         string `form:"estimated_amount"`
	CompanyName          string `form:"company_name"`
	CompanyAddress       string `form:"company_address"`
	ContactPersonName    string `form:"contact_person_name"`
	ConatactPersonMobile string `form:"contact_person_mobile"`
	validator.Validator  `form:"-"`
}

func (app *App) new_project(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	if user.IsAnonymous() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if user.Role != models.FacultyUser {
		app.notFound(w)
		return
	}
	data := app.newTemplateData(r)
	data.Form = newProjectForm{}
	app.render(w, http.StatusOK, "new_project.tmpl", data)
}

func (app *App) new_project_post(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	if user.IsAnonymous() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if user.Role != models.FacultyUser {
		app.notFound(w)
		return
	}

	var form newProjectForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusUnprocessableEntity)
		return
	}

	estimated_amount, err := strconv.Atoi(form.EstimatedAmt)
	form.CheckField(err == nil, "estimated_amount", "Amount must be a number")
	form.CheckField(estimated_amount > 0, "estimated_amount", "This field must be greater than 0")

	form.CheckField(validator.NotBlank(form.CompanyName), "company_name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.CompanyAddress), "company_address", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.ContactPersonName), "contact_person_name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.ConatactPersonMobile), "contact_person_mobile", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "new_project.tmpl", data)
		return
	}

}
