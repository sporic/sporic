package main

import (
	"fmt"
	"math"
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

	if action == "mark_project_completed" {
		sporic_ref_no := r.PostForm.Get("sporic_ref_no")
		err := app.applications.SetStatus(sporic_ref_no, models.ProjectClosed)
		if err != nil {
			app.serverError(w, err)
			return
		}

		var notification models.Notification

		notification.CreatedAt = time.Now()
		notification.NotiType = models.ProjectClosedNotification
		notification.Description = fmt.Sprintf(models.NotificationTypeMap[models.ProjectClosedNotification], sporic_ref_no)
		admins, err := app.users.GetAdmins()
		if err != nil {
			app.serverError(w, err)
			return
		}
		accounts, err := app.users.GetAccounts()
		if err != nil {
			app.serverError(w, err)
			return
		}
		application, err := app.applications.FetchByRefNo(sporic_ref_no)
		if err != nil {
			app.serverError(w, err)
			return
		}
		faculty := strconv.Itoa(application.Leader)
		recievers := append(accounts, admins...)
		recievers = append(recievers, faculty)
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

	for i, application := range applications {

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
			TotalTax += payment.Tax * payment.Payment_amt / 100
		}
		applications[i].TotalAmount = TotalAmt
		applications[i].Taxes = TotalTax

		applications[i].TotalAmountIncludeTax = TotalAmt + TotalTax

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
		applications[i].TotalExpenditure = TotalExpenditure
		applications[i].BalanceAmount = TotalAmt - TotalExpenditure

		var members []models.Member
		members, err = app.applications.GetTeamByRefNo(application.SporicRefNo)
		if err != nil {
			app.serverError(w, err)
			return
		}

		var total_share int
		var total_share_amt float64

		if application.ResourceUsed == 1 {
			total_share_amt = float64(applications[i].BalanceAmount) * 0.6
		} else {
			total_share_amt = float64(applications[i].BalanceAmount) * 0.7
		}
		for i, member := range members {
			share := member.Share
			members[i].MemberShareAmt = int(math.Ceil(total_share_amt * float64(share) / 100))
			total_share += share
		}

		applications[i].LeaderShare = 100 - total_share
		applications[i].LeaderShareAmt = int(math.Ceil(total_share_amt * float64(applications[i].LeaderShare) / 100))

		applications[i].MembersInfo = members

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
