package main

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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
	EndDate                  string   `form:"end_date"`
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
	form.CheckField(validator.NotBlank(form.ProjectTitle), "project_title", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.CompanyName), "company_name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.CompanyAddress), "company_address", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.BillingAddress), "billing_address", "This field cannot be blank.If billing address is same as company address, please enter the same address")
	form.CheckField(validator.NotBlank(form.ContactPersonName), "contact_person_name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.ConatactPersonMobile), "contact_person_mobile", "This field cannot be blank")
	form.CheckField(len(form.ConatactPersonMobile) == 10, "contact_person_mobile", "Enter valid 10-digit contact number")
	form.CheckField(validator.NotBlank(form.ContactPersonEmail), "contact_person_email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.ContactPersonEmail, regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)), "contact_person_email", "Enter a valid email address")
	end_date, err := time.Parse("2006-01-02", form.EndDate)
	form.CheckField(err == nil, "end_date", "Enter a valid end date")
	form.CheckField(end_date.After(time.Now()), "end_date", "Enter a valid end date")

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
	application.EndDate = end_date

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
	err = app.notifications.SendNotification(notification, app.mailer)
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

	if application.Leader != user.Id {
		app.notFound(w)
		return
	}

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
		InvoiceForm, err := app.parseInvoiceForm(r)
		if err != nil {
			app.serverError(w, err)
			return
		}
		if !InvoiceForm.Valid() {
			fmt.Println(InvoiceForm.FieldErrors)
			app.renderFacultyViewApplication(w, r, InvoiceForm, refno)
			return
		}
		err = app.request_invoice(r, *InvoiceForm, application.SporicRefNo)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	if action == "add_expenditure" {
		ExpenditureForm, err := app.parseExpenditureForm(r)
		if err != nil {
			app.serverError(w, err)
			return
		}
		if !ExpenditureForm.Valid() {
			fmt.Println(ExpenditureForm.FieldErrors)
			app.renderFacultyViewApplication(w, r, ExpenditureForm, refno)
			return
		}
		err = app.add_expenditure(r, *ExpenditureForm, application.SporicRefNo)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	if action == "update_payment" {
		UpadatePaymentForm, err := app.parseUpdatePaymentForm(r)
		if err != nil {
			app.serverError(w, err)
			return
		}
		if !UpadatePaymentForm.Valid() {
			fmt.Println(UpadatePaymentForm.FieldErrors)
			app.renderFacultyViewApplication(w, r, UpadatePaymentForm, refno)
			return
		}
		err = app.update_payment(r, *UpadatePaymentForm, application.SporicRefNo)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	if action == "complete_project" {
		closable := true
		for _, payment := range application.Payments {
			if payment.Payment_status != models.PaymentApproved && payment.Payment_status != models.PaymentRejected {
				closable = false
			}
		}
		for _, expenditure := range application.Expenditures {
			if expenditure.Expenditure_status != models.ExpenditureCompleted {
				closable = false
			}
		}
		if closable {
			CompleteProjectForm, err := app.parseCompleteProjectForm(r)
			if err != nil {
				app.serverError(w, err)
				return
			}
			if !CompleteProjectForm.Valid() {
				fmt.Println(CompleteProjectForm.FieldErrors)
				app.renderFacultyViewApplication(w, r, CompleteProjectForm, refno)
				return
			}
			err = app.complete_project(r, *CompleteProjectForm, application.SporicRefNo)
			if err != nil {
				app.serverError(w, err)

				return
			}
		}
	}

	app.renderFacultyViewApplication(w, r, EmptyForm{}, refno)

}

type EmptyForm struct {
	validator.Validator
}

func (app *App) renderFacultyViewApplication(w http.ResponseWriter, r *http.Request, form interface{}, refno string) {
	application, err := app.applications.FetchByRefNo(refno)
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
		return
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
		return
	}

	var total_share int

	for _, member := range members {
		share := member.Share
		total_share += share
	}

	application.LeaderShare = 100 - total_share

	data := app.newTemplateData(r)
	data.Member = members

	data.Form = form
	data.Application = application

	if application.Status == models.ProjectCompleteApprovalPending || application.Status == models.ProjectCompleted {
		app.render(w, http.StatusOK, "faculty_view_completed_application.tmpl", data)
	} else {
		app.render(w, http.StatusOK, "faculty_view_application.tmpl", data)
	}
}

type NewInvoice struct {
	Currency            string `form:"currency"`
	PaymentAmt          int    `form:"payment_amt"`
	Tax                 int    `form:"tax"`
	GstNumber           string `form:"gst_number"`
	PanNumber           string `form:"pan_number"`
	validator.Validator `form:"-"`
}

func (app *App) parseInvoiceForm(r *http.Request) (*NewInvoice, error) {
	var invoice_form NewInvoice

	err := app.decodePostForm(r, &invoice_form, r.PostForm)
	if err != nil {
		return nil, err
	}

	if invoice_form.Currency != "INR" && invoice_form.Currency != "USD" {
		invoice_form.AddFieldError("currency", "Select a valid currency")
	}

	payment_amt := strconv.Itoa(invoice_form.PaymentAmt)
	paymentAmt, err := strconv.Atoi(payment_amt)
	invoice_form.CheckField(err == nil, "payment_amt", "Amount must be a number")
	invoice_form.CheckField(paymentAmt > 0, "payment_amt", "This field must be greater than 0")

	tax := strconv.Itoa(invoice_form.Tax)
	Tax, err := strconv.Atoi(tax)
	invoice_form.CheckField(err == nil, "tax", "Amount must be a number")
	invoice_form.CheckField(Tax > 0, "tax", "This field must be greater than 0")

	invoice_form.CheckField(regexp.MustCompile("^[a-zA-Z0-9]*$").MatchString(invoice_form.GstNumber), "gst_number", "Enter a valid GST number")
	invoice_form.CheckField(len(invoice_form.GstNumber) == 15 || len(invoice_form.GstNumber) == 0, "gst_number", "Enter a valid GST number")

	invoice_form.CheckField(regexp.MustCompile("^[a-zA-Z0-9]*$").MatchString(invoice_form.GstNumber), "pan_number", "Enter a valid PAN number")
	invoice_form.CheckField(len(invoice_form.PanNumber) == 10 || len(invoice_form.PanNumber) == 0, "pan_number", "Enter a valid PAN number")

	return &invoice_form, nil
}

func (app *App) request_invoice(r *http.Request, invoice_form NewInvoice, SporicRefNo string) error {
	var payment models.Payment

	payment.Sporic_ref_no = SporicRefNo
	payment.Currency = invoice_form.Currency
	payment.Payment_amt = invoice_form.PaymentAmt
	payment.Tax = invoice_form.Tax
	payment.Gst_number = strings.ToUpper(invoice_form.GstNumber)
	payment.Pan_number = strings.ToUpper(invoice_form.PanNumber)
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

	err = app.notifications.SendNotification(notification, app.mailer)
	if err != nil {
		return err
	}
	return nil
}

type NewExpenditure struct {
	ExpenditureType     int    `form:"expenditure_type"`
	ExpenditureName     string `form:"expenditure_name"`
	ExpenditureAmt      string `form:"expenditure_amt"`
	validator.Validator `form:"-"`
}

func (app *App) parseExpenditureForm(r *http.Request) (*NewExpenditure, error) {
	var expenditure_form NewExpenditure

	err := app.decodePostForm(r, &expenditure_form, r.PostForm)
	if err != nil {
		return nil, err
	}

	expenditure_amt, err := strconv.Atoi(expenditure_form.ExpenditureAmt)
	expenditure_form.CheckField(err == nil, "expenditure_amt", "Amount must be a number")
	expenditure_form.CheckField(expenditure_amt > 0, "expenditure_amt", "This field must be greater than 0")

	return &expenditure_form, nil
}
func (app *App) add_expenditure(r *http.Request, expenditure_form NewExpenditure, SporicRefNo string) error {
	var expenditure models.Expenditure

	expenditure.SporicRefNo = SporicRefNo
	expenditure.Expenditure_type = expenditure_form.ExpenditureType
	expenditure.Expenditure_name = expenditure_form.ExpenditureName
	expenditure.Expenditure_date = time.Now()
	expenditure.Expenditure_amt, _ = strconv.Atoi(expenditure_form.ExpenditureAmt)
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

	err = app.notifications.SendNotification(notification, app.mailer)
	if err != nil {
		return err
	}

	return nil
}

type CompleteProjectForm struct {
	ResourceUsed        int               `form:"resource_used"`
	Comments            string            `form:"comments"`
	CompletionDate      time.Time         `form:"completion_date"`
	LeaderShare         string            `form:"leader_share"`
	MemberShare         map[string]string `form:"share_percent"`
	validator.Validator `form:"-"`
}

func (app *App) parseCompleteProjectForm(r *http.Request) (*CompleteProjectForm, error) {

	var completion_form CompleteProjectForm

	err := app.decodePostForm(r, &completion_form, r.PostForm)
	if err != nil {
		return nil, err
	}
	completion_form.CheckField((completion_form.ResourceUsed == 1 || completion_form.ResourceUsed == 0), "college_resources", "please enter a valid option")
	completion_form.CheckField(validator.NotBlank(completion_form.Comments), "comments", "cannot leave this field blank")
	leader_share, err := strconv.Atoi(completion_form.LeaderShare)
	completion_form.CheckField(err == nil, "leader_share", "share needs to be a number")
	completion_form.CheckField((leader_share >= 0 && leader_share <= 100), "leader_share", "share needs to be within 0 and 100")
	for _, member_share := range completion_form.MemberShare {
		member_share, err := strconv.Atoi(member_share)
		completion_form.CheckField(err != nil, "member_share", "share needs to be a number")
		completion_form.CheckField((member_share >= 0 && member_share <= 100), "share_percent", "share needs to be within 0 and 100")
	}

	return &completion_form, nil
}

func (app *App) complete_project(r *http.Request, completion_form CompleteProjectForm, SporicRefNo string) error {

	var completion models.Completion
	completion.SporicRefNo = SporicRefNo
	var err error
	completion.LeaderShare, err = strconv.Atoi(completion_form.LeaderShare)
	if err != nil {
		fmt.Println(err)
		return err
	}
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

	err = app.notifications.SendNotification(notification, app.mailer)
	if err != nil {
		return err
	}

	return nil
}

type UpadatePaymentForm struct {
	Payment_id          int    `form:"payment_id"`
	Transaction_id      string `form:"transaction_id"`
	validator.Validator `form:"-"`
}

func (app *App) parseUpdatePaymentForm(r *http.Request) (*UpadatePaymentForm, error) {
	var update_payment_form UpadatePaymentForm

	err := app.decodePostForm(r, &update_payment_form, r.PostForm)

	if err != nil {
		return nil, err
	}

	return &update_payment_form, nil
}
func (app *App) update_payment(r *http.Request, update_payment_form UpadatePaymentForm, SporicRefNo string) error {

	var payment models.Payment

	payment.Sporic_ref_no = SporicRefNo //TODO verify if these 2 belong to the same user or not
	payment.Payment_id = update_payment_form.Payment_id
	payment.Transaction_id = update_payment_form.Transaction_id

	err := app.applications.UpdatePayment(payment)
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

	err = app.notifications.SendNotification(notification, app.mailer)
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
		return
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

		err = app.notifications.SendNotification(notification, app.mailer)
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
		err = app.notifications.SendNotification(notification, app.mailer)
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
		err = app.notifications.SendNotification(notification, app.mailer)
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
		err = app.notifications.SendNotification(notification, app.mailer)
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
		err = app.notifications.SendNotification(notification, app.mailer)
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
		err = app.notifications.SendNotification(notification, app.mailer)
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
		application, err := app.applications.FetchByRefNo(refno)
		if err != nil {
			app.serverError(w, err)
			return
		}
		receivers := accounts
		receivers = append(receivers, strconv.Itoa(application.Leader))
		notification.To = receivers
		err = app.notifications.SendNotification(notification, app.mailer)
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
		return
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
		return
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
		return
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

	user := app.contextGetUser(r)
	if user.IsAnonymous() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	params := httprouter.ParamsFromContext(r.Context())
	folder := params.ByName("folder")
	doc_id := params.ByName("doc_id")
	doc_type := params.ByName("doc_type")

	if folder == "" || doc_id == "" || doc_type == "" {
		http.Error(w, "File not specified.", http.StatusBadRequest)
		return
	}
	if user.Role == models.FacultyUser {

		application, err := app.applications.FetchByRefNo(folder)
		if err != nil {
			app.serverError(w, err)
			return
		}
		if application.Leader != user.Id {
			app.notFound(w)
			return
		}

		if doc_type == "expenditure_proof" || doc_type == "expenditure_invoice" {
			id, err := strconv.Atoi(doc_id)
			if err != nil {
				app.serverError(w, err)
				return
			}
			expenditure, err := app.applications.GetExpenditureById(id)
			if err != nil {
				app.serverError(w, err)
				return
			}

			if expenditure.SporicRefNo != application.SporicRefNo {
				app.notFound(w)
				return
			}

		}

		if doc_type == "invoice" || doc_type == "payment" || doc_type == "tax_cirtificate" {
			id, err := strconv.Atoi(doc_id)
			if err != nil {
				app.serverError(w, err)
				return
			}
			payment, err := app.payments.GetPaymentById(id)
			if err != nil {
				app.serverError(w, err)
				return
			}

			if payment.Sporic_ref_no != application.SporicRefNo {
				app.notFound(w)
				return
			}
		}

		if application.SporicRefNo != folder {
			app.notFound(w)
			return
		}
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
			return
		}

		admins, err := app.users.GetAdmins()
		if err != nil {
			app.serverError(w, err)
			return
		}
		application, err := app.applications.FetchByRefNo(sporic_ref_no)
		if err != nil {
			app.serverError(w, err)
			return
		}
		user := strconv.Itoa(application.Leader)
		var notification models.Notification
		notification.CreatedAt = time.Now()
		notification.NotiType = models.InvoiceUploaded
		notification.Description = fmt.Sprintf(models.NotificationTypeMap[models.InvoiceUploaded], payment_id, sporic_ref_no)
		receivers := admins
		receivers = append(receivers, user)
		notification.To = receivers

		err = app.notifications.SendNotification(notification, app.mailer)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	if action == "approve_payment" {
		payment_id := r.PostForm.Get("payment_id")
		sporic_ref_no := r.PostForm.Get("sporic_ref_no")
		err := app.applications.SetPaymentStatus(payment_id, models.PaymentApproved)
		if err != nil {
			app.serverError(w, err)
			return
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
		err = app.notifications.SendNotification(notification, app.mailer)
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
		err = app.notifications.SendNotification(notification, app.mailer)
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
		err = app.notifications.SendNotification(notification, app.mailer)
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
	if user.Role != models.AdminUser {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}

	from_date := r.PostForm.Get("from_date")
	to_date := r.PostForm.Get("to_date")

	if from_date == "" || to_date == "" {
		http.Error(w, "Please enter both dates", http.StatusBadRequest)
		return
	}

	fromDate, err := time.Parse("2006-01-02", from_date)
	if err != nil {
		app.serverError(w, err)
		return
	}

	toDate, err := time.Parse("2006-01-02", to_date)
	if err != nil {
		app.serverError(w, err)
		return
	}

	applications, err := app.applications.FetchAll()
	if err != nil {
		app.serverError(w, err)
		return
	}

	var filtered_applications []models.Application

	for _, application := range applications {
		if application.StartDate.After(fromDate) && application.StartDate.Before(toDate) {
			filtered_applications = append(filtered_applications, application)
		}
	}

	file, err := app.GenerateExcel(filtered_applications)
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

			err = app.notifications.SendNotification(notification, app.mailer)
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

				err = app.notifications.SendNotification(notification, app.mailer)
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

func (app *App) profile(w http.ResponseWriter, r *http.Request) {

	user := app.contextGetUser(r)

	data := app.newTemplateData(r)
	data.User = user
	app.render(w, http.StatusOK, "profile.tmpl", data)
}
