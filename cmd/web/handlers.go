package main

import (
	"fmt"
	"net/http"
	"path/filepath"
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
	if user.Role == models.Provc {
		http.Redirect(w, r, "/provc_home", http.StatusSeeOther)
	}
	app.notFound(w)
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
