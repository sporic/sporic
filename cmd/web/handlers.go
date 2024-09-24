package main

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/sporic/sporic/internal/models"
	"github.com/sporic/sporic/internal/validator"
)

type FileType = int

const (
	ProposalDoc FileType = iota
	Invoice
	PaymentProof
	GstCirtificate
	PanCard
	CompletionDoc
	ExpenditureProof
	ExpenditureInvoice
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
	if user.Role == models.AccountantUser {
		http.Redirect(w, r, "/accounts_home", http.StatusSeeOther)
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
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	applications, err := app.applications.FetchAll()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Applications = applications
	app.render(w, http.StatusOK, "admin_home.tmpl", data)
}

func (app *App) faculty_home(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	if user.IsAnonymous() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if user.Role != models.FacultyUser {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	applications, err := app.applications.FetchByLeader(user.Id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Applications = applications
	app.render(w, http.StatusOK, "faculty_home.tmpl", data)
}

type newApplicationForm struct {
	ProjectTitle             string   `form:"project_title"`
	ActivityType             string   `form:"activity_type"`
	FinancialYear            string   `form:"financial_year"`
	EstimatedAmt             string   `form:"estimated_amount"`
	CompanyName              string   `form:"company_name"`
	CompanyAddress           string   `form:"company_address"`
	BillingAddress           string   `form:"billing_address"`
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

	err := r.ParseMultipartForm(10 << 20)
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

	application.ProjectTitle = form.ProjectTitle
	application.FinancialYear = fy_year
	application.EstimatedAmt = estimated_amount
	application.CompanyName = form.CompanyName
	application.CompanyAddress = form.CompanyAddress
	application.BillingAddress = form.BillingAddress
	application.ContactPersonName = form.ContactPersonName
	application.ContactPersonEmail = form.ContactPersonEmail
	application.ContactPersonMobile = form.ConatactPersonMobile
	application.ContactPersonDesignation = form.ContactPersonDesignation
	application.Members = form.Members

	sporic_ref_no, err := app.applications.Insert(application)
	if err != nil {
		app.serverError(w, err)
		return
	}
	err = app.handleFile(r, sporic_ref_no, sporic_ref_no, ProposalDoc, "project_proposal")
	if err != nil {
		fmt.Println(err)
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
		return
	}

	if r.Method == http.MethodPost {
		err = r.ParseMultipartForm(10 << 20)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
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

	if action == "update_payment" {
		err = app.update_payment(r, application.SporicRefNo)
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
	Currency   string `form:"currency"`
	PaymentAmt int    `form:"payment_amt"`
	Tax        int    `form:"tax"`
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
	payment.Currency = invoice_form.Currency
	payment.Payment_amt = invoice_form.PaymentAmt
	payment.Tax = invoice_form.Tax
	payment.Gst_number = invoice_form.GstNumber
	payment.Pan_number = invoice_form.PanNumber
	payment.Payment_status = models.PaymentInvoiceRequested

	id, err := app.applications.Insert_invoice_request(payment)
	if err != nil {
		return err
	}

	err = app.handleFile(r, SporicRefNo, strconv.Itoa(id), GstCirtificate, "tax_certificate")
	if err != nil {
		return err
	}

	return nil
}

type NewExpenditure struct {
	ExpenditureType int    `form:"expenditure_type"`
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
	expenditure.Expenditure_type = expenditure_form.ExpenditureType
	expenditure.Expenditure_name = expenditure_form.ExpenditureName
	expenditure.Expenditure_date = time.Now()
	expenditure.Expenditure_amt = expenditure_form.ExpenditureAmt
	expenditure.Expenditure_status = models.ExpenditurePendingApproval

	exp_id, err := app.applications.Insert_expenditure(expenditure)
	if err != nil {
		return err
	}

	err = app.handleFile(r, SporicRefNo, strconv.Itoa(exp_id), ExpenditureProof, "expenditure_proof")

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

	err = app.handleFile(r, SporicRefNo, SporicRefNo, CompletionDoc, "project_closure_report")

	if err != nil {
		return err
	}

	return nil
}

type UpadatePaymentForm struct {
	Payment_id     int    `form:"payment_id"`
	Transaction_id string `form:"transaction_id"`
}

func (app *App) update_payment(r *http.Request, SporicRefNo string) error {

	var update_payment_form UpadatePaymentForm

	err := app.decodePostForm(r, &update_payment_form, r.PostForm)

	if err != nil {
		return err
	}

	var payment models.Payment

	payment.Sporic_ref_no = SporicRefNo
	payment.Payment_id = update_payment_form.Payment_id
	payment.Transaction_id = update_payment_form.Transaction_id

	err = app.applications.UpdatePayment(payment)
	if err != nil {
		return err
	}

	err = app.handleFile(r, SporicRefNo, strconv.Itoa(payment.Payment_id), PaymentProof, "payment_proof")
	if err != nil {
		return err
	}

	return nil
}

func (app *App) admin_view_application(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	if user.IsAnonymous() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if user.Role != models.AdminUser {
		app.notFound(w)
		return
	}
	params := httprouter.ParamsFromContext(r.Context())
	refno := params.ByName("refno")
	_, err := app.applications.FetchByRefNo(refno)
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

	if action == "approve_completion" {
		err = app.applications.SetStatus(refno, models.ProjectCompleted)
		if err != nil {
			app.serverError(w, err)

			return
		}
	}
	if action == "approve_application" {
		err = app.applications.SetStatus(refno, models.ProjectApproved)
		if err != nil {
			app.serverError(w, err)

			return
		}
	}
	if action == "reject_application" {
		err = app.applications.SetStatus(refno, models.ProjectRejected)
		if err != nil {
			app.serverError(w, err)

			return
		}
	}
	if action == "approve_expenditure" {
		err = app.applications.SetExpenditureStatus(refno, models.ExpenditureApproved)
		if err != nil {
			app.serverError(w, err)

			return
		}
	}
	if action == "reject_expenditure" {
		err = app.applications.SetExpenditureStatus(refno, models.ExpenditureRejected)
		if err != nil {
			app.serverError(w, err)

			return
		}
	}
	if action == "invoice_forwared" {
		err = app.applications.SetPaymentStatus(refno, models.PaymentInvoiceForwarded)
		if err != nil {
			app.serverError(w, err)

			return
		}

	}

	application, err := app.applications.FetchByRefNo(refno)

	if errors.Is(err, models.ErrRecordNotFound) {
		app.notFound(w)
		return
	}
	if err != nil {
		app.serverError(w, err)
		return
	}

	user, err = app.users.Get(application.Leader)
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
	data.User = user

	app.render(w, http.StatusOK, "admin_view_application.tmpl", data)

}

func (app *App) download(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	folder := params.ByName("folder")
	doc_id := params.ByName("doc_id")
	doc_type := params.ByName("doc_type")

	fmt.Println(folder, doc_id, doc_type)

	if folder == "" || doc_id == "" || doc_type == "" {
		http.Error(w, "File not specified.", http.StatusBadRequest)
		return
	}

	filename := folder + "_" + doc_id + "_" + doc_type + ".pdf"

	prefixPath := "Documents/" + folder + "/"

	filePath := filepath.Join(prefixPath, filepath.Clean(filename))

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, filePath)
}

func (app *App) accounts_home(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	if user.IsAnonymous() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if user.Role != models.AccountantUser {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	payments, err := app.payments.GetAllPayments()
	if err != nil {
		app.serverError(w, err)
	}

	data := app.newTemplateData(r)
	data.User = user
	data.Payments = payments
	app.render(w, http.StatusOK, "accounts_home.tmpl", data)
}
