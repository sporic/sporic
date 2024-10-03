package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sporic/sporic/internal/models"
)

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
		err := app.applications.SetPaymentStatus(payment_id, models.PaymentCompleted)
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
