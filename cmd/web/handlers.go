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
	FeedbackForm
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
	MemberStudents           []string `form:"member_students"`
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
	application.MemberStudents = form.MemberStudents

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

	var notification models.Notification

	notification.CreatedAt = time.Now()
	notification.NotiType = models.NewProjectApproval
	notification.Description = fmt.Sprintf(models.NotificationTypeMap[models.NewProjectApproval], sporic_ref_no)
	admins, err := app.users.GetAdmins()
	if err != nil {
		app.serverError(w, err)
		return
	}
	notification.To = admins
	err = app.notifications.SendNotification(notification)
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

	var payments []models.Payment
	payments, err = app.payments.GetPaymentByRefNo(application.SporicRefNo)
	if err != nil {
		app.serverError(w, err)
	}
	var TotalAmt int = 0
	var TotalTax int = 0
	for _, payment := range payments {
		TotalAmt += payment.Payment_amt
		TotalTax += payment.Tax
	}
	application.TotalAmount = TotalAmt
	application.Taxes = TotalTax

	var expenditures []models.Expenditure
	expenditures, err = app.applications.GetExpenditureByRefNo(application.SporicRefNo)
	if err != nil {
		app.serverError(w, err)
	}
	var TotalExpenditure int = 0
	for _, expenditure := range expenditures {
		TotalExpenditure += expenditure.Expenditure_amt
	}
	application.TotalExpenditure = TotalExpenditure
	application.BalanceAmount = TotalAmt - TotalExpenditure

	var members []models.Member
	members, err = app.applications.GetTeamByRefNo(application.SporicRefNo)
	if err != nil {
		app.serverError(w, err)
	}

	var total_share int

	for _, member := range members {
		share := member.Share
		total_share += share
	}

	application.LeaderShare = 100 - total_share

	data := app.newTemplateData(r)
	data.Member = members
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

	var notification models.Notification

	notification.CreatedAt = time.Now()
	notification.NotiType = models.NewInvoiceRequestApproval
	notification.Description = fmt.Sprintf(models.NotificationTypeMap[models.NewInvoiceRequestApproval], SporicRefNo)

	admins, err := app.users.GetAdmins()
	if err != nil {
		return err
	}
	notification.To = admins

	err = app.notifications.SendNotification(notification)
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

	if expenditure_form.ExpenditureType == 0 {
		err = app.handleFile(r, SporicRefNo, strconv.Itoa(exp_id), ExpenditureProof, "expenditure_proof")
	}
	if expenditure_form.ExpenditureType == 1 {
		err = app.handleFile(r, SporicRefNo, strconv.Itoa(exp_id), ExpenditureInvoice, "expenditure_proof")
	}

	if err != nil {
		return err
	}

	var notification models.Notification

	notification.CreatedAt = time.Now()
	notification.NotiType = models.NewExpenditureApproval
	notification.Description = fmt.Sprintf(models.NotificationTypeMap[models.NewExpenditureApproval], SporicRefNo)

	admins, err := app.users.GetAdmins()
	if err != nil {
		return err
	}
	notification.To = admins

	err = app.notifications.SendNotification(notification)
	if err != nil {
		return err
	}

	return nil
}

type CompleteProjectForm struct {
	ResourceUsed   int               `form:"resource_used"`
	Comments       string            `form:"comments"`
	CompletionDate time.Time         `form:"completion_date"`
	LeaderShare    int               `form:"leader_share"`
	MemberShare    map[string]string `form:"share_percent"`
}

func (app *App) complete_project(r *http.Request, SporicRefNo string) error {

	var completion_form CompleteProjectForm

	err := app.decodePostForm(r, &completion_form, r.PostForm)
	if err != nil {
		return err
	}

	var completion models.Completion
	completion.SporicRefNo = SporicRefNo
	completion.LeaderShare = completion_form.LeaderShare
	completion.MemberShare = completion_form.MemberShare
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

	err = app.handleFile(r, SporicRefNo, SporicRefNo, FeedbackForm, "feedback_form")

	if err != nil {
		return err
	}

	var notification models.Notification

	notification.CreatedAt = time.Now()
	notification.NotiType = models.CompletionProjectApproval
	notification.Description = fmt.Sprintf(models.NotificationTypeMap[models.CompletionProjectApproval], SporicRefNo)

	admins, err := app.users.GetAdmins()
	if err != nil {
		return err
	}
	notification.To = admins

	err = app.notifications.SendNotification(notification)
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

	var notification models.Notification

	notification.CreatedAt = time.Now()
	notification.NotiType = models.PaymentApproval
	notification.Description = fmt.Sprintf(models.NotificationTypeMap[models.PaymentApproval], strconv.Itoa(payment.Payment_id), SporicRefNo)

	accounts, err := app.users.GetAccounts()
	if err != nil {
		return err
	}
	notification.To = accounts

	err = app.notifications.SendNotification(notification)
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

	if action == "approve_completion" {
		err = app.applications.SetStatus(refno, models.ProjectCompleted)
		if err != nil {
			app.serverError(w, err)
			return
		}

		var notification models.Notification

		notification.CreatedAt = time.Now()
		notification.NotiType = models.ApplicationCompleted
		notification.Description = fmt.Sprintf(models.NotificationTypeMap[models.ApplicationCompleted], refno)
		var recievers []string
		recievers = append(recievers, strconv.Itoa(application.Leader))
		notification.To = recievers

		err = app.notifications.SendNotification(notification)
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

		var notification models.Notification

		notification.CreatedAt = time.Now()
		notification.NotiType = models.ApplicationApproved
		notification.Description = fmt.Sprintf(models.NotificationTypeMap[models.ApplicationApproved], refno)
		var recievers []string
		recievers = append(recievers, strconv.Itoa(application.Leader))
		notification.To = recievers
		err = app.notifications.SendNotification(notification)
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

		var notification models.Notification
		notification.CreatedAt = time.Now()
		notification.NotiType = models.ApplicationRejected
		notification.Description = fmt.Sprintf(models.NotificationTypeMap[models.ApplicationRejected], refno)
		var recievers []string
		recievers = append(recievers, strconv.Itoa(application.Leader))
		notification.To = recievers
		err = app.notifications.SendNotification(notification)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}
	if action == "approve_expenditure" {
		expenditure_id := r.PostForm.Get("expenditure_id")
		err = app.applications.SetExpenditureStatus(expenditure_id, models.ExpenditureApproved)
		if err != nil {
			app.serverError(w, err)
			return
		}

		var notification models.Notification
		notification.CreatedAt = time.Now()
		notification.NotiType = models.ExpenditureApprovedNotification
		notification.Description = fmt.Sprintf(models.NotificationTypeMap[models.ExpenditureApprovedNotification], refno)
		var recievers []string
		recievers = append(recievers, strconv.Itoa(application.Leader))
		notification.To = recievers
		err = app.notifications.SendNotification(notification)
		if err != nil {
			app.serverError(w, err)
			return
		}

		notification.CreatedAt = time.Now()
		notification.NotiType = models.ExpenditureApprovedNotification
		notification.Description = fmt.Sprintf(models.NotificationTypeMap[models.ExpenditureApprovedNotification], refno)

		accounts, err := app.users.GetAccounts()
		if err != nil {
			app.serverError(w, err)
			return
		}
		notification.To = accounts
		err = app.notifications.SendNotification(notification)
		if err != nil {
			app.serverError(w, err)
			return
		}

	}
	if action == "reject_expenditure" {
		expenditure_id := r.PostForm.Get("expenditure_id")
		err = app.applications.SetExpenditureStatus(expenditure_id, models.ExpenditureRejected)
		if err != nil {
			app.serverError(w, err)
			return
		}

		var notification models.Notification
		notification.CreatedAt = time.Now()
		notification.NotiType = models.ExpenditureRejectedNotification
		notification.Description = fmt.Sprintf(models.NotificationTypeMap[models.ExpenditureRejectedNotification], refno)
		var recievers []string
		recievers = append(recievers, strconv.Itoa(application.Leader))
		notification.To = recievers
		err = app.notifications.SendNotification(notification)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}
	if action == "invoice_forwared" {
		payment_id := r.PostForm.Get("payment_id")
		err = app.applications.SetPaymentStatus(payment_id, models.PaymentInvoiceForwarded)
		if err != nil {
			app.serverError(w, err)
			return
		}
		var notification models.Notification
		notification.CreatedAt = time.Now()
		notification.NotiType = models.PaymentInvoiceRequest
		notification.Description = fmt.Sprintf(models.NotificationTypeMap[models.PaymentInvoiceRequest], refno)

		accounts, err := app.users.GetAccounts()
		if err != nil {
			app.serverError(w, err)
			return
		}
		notification.To = accounts
		err = app.notifications.SendNotification(notification)
		if err != nil {
			app.serverError(w, err)
			return
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

	user, err = app.users.Get(application.Leader)
	if errors.Is(err, models.ErrRecordNotFound) {
		app.notFound(w)
		return
	}
	if err != nil {
		app.serverError(w, err)
		return
	}

	var payments []models.Payment
	payments, err = app.payments.GetPaymentByRefNo(application.SporicRefNo)
	if err != nil {
		app.serverError(w, err)
	}
	var TotalAmt int = 0
	var TotalTax int = 0
	for _, payment := range payments {
		TotalAmt += payment.Payment_amt
		TotalTax += payment.Tax
	}
	application.TotalAmount = TotalAmt
	application.Taxes = TotalTax

	var expenditures []models.Expenditure
	expenditures, err = app.applications.GetExpenditureByRefNo(application.SporicRefNo)
	if err != nil {
		app.serverError(w, err)
	}
	var TotalExpenditure int = 0
	for _, expenditure := range expenditures {
		TotalExpenditure += expenditure.Expenditure_amt
	}
	application.TotalExpenditure = TotalExpenditure
	application.BalanceAmount = TotalAmt - TotalExpenditure

	var members []models.Member
	members, err = app.applications.GetTeamByRefNo(application.SporicRefNo)
	if err != nil {
		app.serverError(w, err)
	}

	var total_share int

	for _, member := range members {
		share := member.Share
		total_share += share
	}

	application.LeaderShare = 100 - total_share

	data := app.newTemplateData(r)
	data.Member = members
	data.Application = application
	data.User = user

	app.render(w, http.StatusOK, "admin_view_application.tmpl", data)
}

func (app *App) download(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	folder := params.ByName("folder")
	doc_id := params.ByName("doc_id")
	doc_type := params.ByName("doc_type")

	if folder == "" || doc_id == "" || doc_type == "" {
		http.Error(w, "File not specified.", http.StatusBadRequest)
		return
	}

	filename := folder + "_" + doc_id + "_" + doc_type + ".pdf"

	prefixPath := "Documents/" + folder + "/"

	filePath := filepath.Join(prefixPath, filepath.Clean(filename))

	w.Header().Set("Content-Type", "application/pdf")

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

	var action string
	if r.Method == http.MethodPost {
		err := r.ParseMultipartForm(10 << 20)

		if err != nil {
			app.clientError(w, http.StatusBadRequest)
		}
	}

	action = r.PostForm.Get("action")

	if action == "generate_invoice" {
		payment_id := r.PostForm.Get("payment_id")
		sporic_ref_no := r.PostForm.Get("sporic_ref_no")
		err := app.UploadInvoice(r, payment_id, sporic_ref_no)
		if err != nil {
			app.serverError(w, err)
		}
	}

	if action == "approve_payment" {
		payment_id := r.PostForm.Get("payment_id")
		sporic_ref_no := r.PostForm.Get("sporic_ref_no")
		err := app.applications.SetPaymentStatus(payment_id, models.PaymentApproved)
		if err != nil {
			app.serverError(w, err)
		}

		var notification models.Notification
		notification.CreatedAt = time.Now()
		notification.NotiType = models.PaymentApprovedNotification
		notification.Description = fmt.Sprintf(models.NotificationTypeMap[models.PaymentApprovedNotification], payment_id, sporic_ref_no)
		application, err := app.applications.FetchByRefNo(sporic_ref_no)
		if err != nil {
			app.serverError(w, err)
			return
		}
		var recievers []string
		recievers = append(recievers, strconv.Itoa(application.Leader))
		notification.To = recievers
		err = app.notifications.SendNotification(notification)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	if action == "reject_payment" {
		payment_id := r.PostForm.Get("payment_id")
		sporic_ref_no := r.PostForm.Get("sporic_ref_no")
		err := app.applications.SetPaymentStatus(payment_id, models.PaymentRejected)
		if err != nil {
			app.serverError(w, err)
			return
		}

		var notification models.Notification
		notification.CreatedAt = time.Now()
		notification.NotiType = models.PaymentRejectedNotification
		notification.Description = fmt.Sprintf(models.NotificationTypeMap[models.PaymentRejectedNotification], payment_id, sporic_ref_no)
		application, err := app.applications.FetchByRefNo(sporic_ref_no)
		if err != nil {
			app.serverError(w, err)
			return
		}
		var recievers []string
		recievers = append(recievers, strconv.Itoa(application.Leader))
		notification.To = recievers
		err = app.notifications.SendNotification(notification)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	if action == "complete_expenditure" {
		expenditure_id := r.PostForm.Get("expenditure_id")
		sporic_ref_no := r.PostForm.Get("sporic_ref_no")
		err := app.applications.SetExpenditureStatus(expenditure_id, models.ExpenditureCompleted)
		if err != nil {
			app.serverError(w, err)
			return
		}
		var notification models.Notification
		notification.CreatedAt = time.Now()
		notification.NotiType = models.ExpenditurePaid
		notification.Description = fmt.Sprintf(models.NotificationTypeMap[models.ExpenditurePaid], expenditure_id, sporic_ref_no)
		application, err := app.applications.FetchByRefNo(sporic_ref_no)
		if err != nil {
			app.serverError(w, err)
			return
		}
		var recievers []string
		admins, err := app.users.GetAdmins()
		if err != nil {
			return
		}
		recievers = admins
		recievers = append(recievers, strconv.Itoa(application.Leader))
		notification.To = recievers
		err = app.notifications.SendNotification(notification)
		if err != nil {
			app.serverError(w, err)
			return
		}

	}

	payments, err := app.payments.GetAllPayments()
	if err != nil {
		app.serverError(w, err)
		return
	}

	applications, err := app.applications.FetchAll()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Applications = applications
	data.User = user
	data.Payments = payments
	app.render(w, http.StatusOK, "accounts_home.tmpl", data)
}

func (app *App) UploadInvoice(r *http.Request, payment_id string, sporic_ref_no string) error {

	err := app.handleFile(r, sporic_ref_no, payment_id, Invoice, "invoice")
	if err != nil {
		return err
	}

	err = app.applications.SetPaymentStatus(payment_id, models.PaymentPending)
	if err != nil {
		return err
	}

	return nil
}

func (app *App) excel(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	if user.IsAnonymous() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if user.Role != models.FacultyUser {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	applications, err := app.applications.FetchAll()
	if err != nil {
		app.serverError(w, err)
	}

	file, err := app.GenerateExcel(applications)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=sporic-applications.xlsx")

	if err := file.Write(w); err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *App) checkForDelayes() {

	var applications []models.Application

	applications, err := app.applications.FetchAll()
	if err != nil {
		app.serverError(nil, err)
		return
	}

	current_time := time.Now()

	for _, application := range applications {
		if application.EndDate.Before(current_time) {
			var notification models.Notification

			notification.CreatedAt = time.Now()
			notification.NotiType = models.ProjectDelayed
			notification.Description = fmt.Sprintf(models.NotificationTypeMap[models.ProjectDelayed], application.SporicRefNo)

			admins, err := app.users.GetAdmins()
			user := strconv.Itoa(application.Leader)
			if err != nil {
				return
			}
			recievers := append(admins, user)
			notification.To = recievers

			err = app.notifications.SendNotification(notification)
			if err != nil {
				return
			}
		}

		for _, payment := range application.Payments {
			if payment.Payment_status == models.PaymentPending && payment.Payment_date.Valid && (current_time.Sub(payment.Payment_date.Time).Hours()/24 >= 90) {
				var notification models.Notification

				notification.CreatedAt = time.Now()
				notification.NotiType = models.PaymentDelayed
				notification.Description = fmt.Sprintf(models.NotificationTypeMap[models.PaymentDelayed], application.SporicRefNo)

				admins, err := app.users.GetAdmins()
				user := strconv.Itoa(application.Leader)
				if err != nil {
					return
				}
				recievers := append(admins, user)
				notification.To = recievers

				err = app.notifications.SendNotification(notification)
				if err != nil {
					return
				}
			}
		}
	}
}

func (app *App) GetNotifications(w http.ResponseWriter, r *http.Request) {

	user := app.contextGetUser(r)
	if user.IsAnonymous() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var notifications []models.Notification
	var err error
	if user.Role == models.AdminUser {

		admins, err := app.users.GetAdmins()
		if err != nil {
			app.serverError(w, err)
			return
		}
		notifications, err = app.notifications.RecieveNotification(admins)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}
	if user.Role == models.FacultyUser {
		var receivers []string

		receivers = append(receivers, strconv.Itoa(user.Id))
		notifications, err = app.notifications.RecieveNotification(receivers)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}
	if user.Role == models.AccountantUser {
		accounts, err := app.users.GetAccounts()
		if err != nil {
			app.serverError(w, err)
			return
		}
		notifications, err = app.notifications.RecieveNotification(accounts)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	data := app.newTemplateData(r)
	data.Notifications = notifications
	app.render(w, http.StatusOK, "notifications.tmpl", data)
}
