package main

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
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

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}

	err = app.decodePostForm(r, &form, r.PostForm)
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
	applications, err := app.applications.FetchAll()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Applications = applications
	app.render(w, http.StatusOK, "faculty_home.tmpl", data)
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
	applications, err := app.applications.FetchByLeader(user.Id)
	fmt.Println(applications)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Applications = applications
	app.render(w, http.StatusOK, "faculty_home.tmpl", data)
}

type newApplicationForm struct {
	ActivityType             string   `form:"activity_type"`
	FinancialYear            string   `form:"financial_year"`
	EstimatedAmt             string   `form:"estimated_amount"`
	CompanyName              string   `form:"company_name"`
	CompanyAddress           string   `form:"company_address"`
	ContactPersonName        string   `form:"contact_person_name"`
	ContactPersonEmail       string   `form:"contact_person_email"`
	ConatactPersonMobile     string   `form:"contact_person_mobile"`
	ContactPersonDesignation string   `form:"contact_person_designation"`
	Members                  []string `form:"members"`
	validator.Validator      `form:"-"`
}

func (app *App) new_application(w http.ResponseWriter, r *http.Request) {
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
	data.Form = newApplicationForm{}
	app.render(w, http.StatusOK, "new_application.tmpl", data)
}

func (app *App) new_application_post(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	if user.IsAnonymous() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if user.Role != models.FacultyUser {
		app.notFound(w)
		return
	}

	err := r.ParseForm()

	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}

	var form newApplicationForm

	err = app.decodePostForm(r, &form, r.PostForm)
	if err != nil {
		app.clientError(w, http.StatusUnprocessableEntity)
		return
	}

	var application models.Application
	application.Leader = user.Id

	switch form.ActivityType {
	case "consultancy":
		application.ActivityType = models.ActivityTypeConsultancy
	case "training":
		application.ActivityType = models.ActivityTypeTraining
	default:
		form.AddFieldError("activity_type", "Select a valid activity type")
	}

	estimated_amount, err := strconv.Atoi(form.EstimatedAmt)
	form.CheckField(err == nil, "estimated_amount", "Amount must be a number")
	form.CheckField(estimated_amount > 0, "estimated_amount", "This field must be greater than 0")

	form.CheckField(validator.NotBlank(form.FinancialYear), "financial_year", "This field cannot be blank")
	form.CheckField(validator.Matches(form.FinancialYear, regexp.MustCompile(`^\d{4}$`)), "financial_year", "This field must be a 4 digit number")
	fy_year, _ := strconv.Atoi(form.FinancialYear)

	form.CheckField(validator.NotBlank(form.CompanyName), "company_name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.CompanyAddress), "company_address", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.ContactPersonName), "contact_person_name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.ConatactPersonMobile), "contact_person_mobile", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.ContactPersonEmail), "contact_person_email", "This field cannot be blank")

	if !form.Valid() {
		fmt.Println(form.FieldErrors)
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "new_application.tmpl", data)
		return
	}

	application.FinancialYear = fy_year
	application.EstimatedAmt = estimated_amount
	application.CompanyName = form.CompanyName
	application.CompanyAddress = form.CompanyAddress
	application.ContactPersonName = form.ContactPersonName
	application.ContactPersonEmail = form.ContactPersonEmail
	application.ContactPersonMobile = form.ConatactPersonMobile
	application.ContactPersonDesignation = form.ContactPersonDesignation
	application.Members = form.Members

	err = app.applications.Insert(application)
	if err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, "/faculty_home", http.StatusSeeOther)
}

func (app *App) faculty_view_application(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	if user.IsAnonymous() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if user.Role != models.FacultyUser {
		app.notFound(w)
		return
	}
	params := httprouter.ParamsFromContext(r.Context())
	refno := params.ByName("refno")
	application, err := app.applications.FetchByRefNo(refno)
	if errors.Is(err, models.ErrRecordNotFound) {
		app.notFound(w)
		return
	} else if err != nil {
		app.serverError(w, err)
	}

	err = r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}

	action := r.PostForm.Get("action")

	if action == "request_invoice" {
		err = app.request_invoice(r, application.SporicRefNo)
		if err != nil {
			app.serverError(w, err)

			return
		}
	}

	if action == "add_expenditure" {
		err = app.add_expenditure(r, application.SporicRefNo)
		if err != nil {
			app.serverError(w, err)

			return
		}
	}

	if action == "complete_project" {
		closable := true
		for _, payment := range application.Payments {
			if payment.Payment_status != models.PaymentApproved {
				closable = false
			}
		}
		for _, expenditure := range application.Expenditures {
			if expenditure.Expenditure_status != models.ExpenditureApproved {
				closable = false
			}
		}
		// TODO remove
		closable = true
		if closable {
			err = app.complete_project(r, application.SporicRefNo)
			if err != nil {
				app.serverError(w, err)

				return
			}
		} else {
			// TODO form error
		}
	}

	application, err = app.applications.FetchByRefNo(refno)
	if errors.Is(err, models.ErrRecordNotFound) {
		app.notFound(w)
		return
	}
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Application = application

	if application.Status == models.ProjectCompleteApprovalPending || application.Status == models.ProjectCompleted {
		app.render(w, http.StatusOK, "faculty_view_completed_application.tmpl", data)
	} else {
		app.render(w, http.StatusOK, "faculty_view_application.tmpl", data)
	}

}

type NewInvoice struct {
	PaymentAmt int    `form:"payment_amt"`
	GstNumber  string `form:"gst_number"`
	PanNumber  string `form:"pan_number"`
}

func (app *App) request_invoice(r *http.Request, SporicRefNo string) error {

	var invoice_form NewInvoice

	err := app.decodePostForm(r, &invoice_form, r.PostForm)
	if err != nil {
		return err
	}

	var payment models.Payment

	payment.Sporic_ref_no = SporicRefNo
	payment.Payment_amt = invoice_form.PaymentAmt
	payment.Gst_number = invoice_form.GstNumber
	payment.Pan_number = invoice_form.PanNumber
	payment.Payment_status = models.PaymentInvoiceRequested

	err = app.applications.Insert_invoice_request(payment)
	if err != nil {
		return err
	}

	return nil
}

type NewExpenditure struct {
	ExpenditureName string `form:"expenditure_name"`
	ExpenditureAmt  int    `form:"expenditure_amt"`
}

func (app *App) add_expenditure(r *http.Request, SporicRefNo string) error {

	var expenditure_form NewExpenditure

	err := app.decodePostForm(r, &expenditure_form, r.PostForm)
	if err != nil {
		return err
	}

	var expenditure models.Expenditure

	expenditure.SporicRefNo = SporicRefNo
	expenditure.Expenditure_name = expenditure_form.ExpenditureName
	expenditure.Expenditure_date = time.Now()
	expenditure.Expenditure_amt = expenditure_form.ExpenditureAmt
	expenditure.Expenditure_status = models.ExpenditurePendingApproval

	err = app.applications.Insert_expenditure(expenditure)
	if err != nil {
		return err
	}

	return nil
}

type CompleteProjectForm struct {
	ResourceUsed   int       `form:"resource_used"`
	Comments       string    `form:"comments"`
	CompletionDate time.Time `form:"completion_date"`
}

func (app *App) complete_project(r *http.Request, SporicRefNo string) error {

	var completion_form CompleteProjectForm

	err := app.decodePostForm(r, &completion_form, r.PostForm)
	if err != nil {
		return err
	}

	var completion models.Completion
	completion.SporicRefNo = SporicRefNo
	completion.ResourceUsed = completion_form.ResourceUsed
	completion.Comments = completion_form.Comments
	completion_form.CompletionDate = time.Now()

	err = app.applications.Complete_Project(completion)
	if err != nil {
		return err
	}

	return nil
}
