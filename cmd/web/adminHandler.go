package main

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/sporic/sporic/internal/models"
)

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

		//err = app.applications.AdminCompleteApplication(refno)

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

		provc, err := app.users.GetProvc()
		if err != nil {
			app.serverError(w, err)
			return
		}

		provc_notifications := models.Notification{
			CreatedAt:   time.Now(),
			NotiType:    models.ProVCApprovalPending,
			Description: fmt.Sprintf(models.NotificationTypeMap[models.ProVCApprovalPending], refno),
			To:          provc,
		}
		err = app.notifications.SendNotification(provc_notifications, app.mailer)
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
		TotalTax += payment.Tax * payment.Payment_amt / 100
	}
	application.TotalAmount = TotalAmt
	application.Taxes = TotalTax

	application.TotalAmountIncludeTax = TotalAmt + TotalTax

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
	var total_share_amt float64

	if application.ResourceUsed == 1 {
		total_share_amt = float64(application.BalanceAmount) * 0.6
	} else {
		total_share_amt = float64(application.BalanceAmount) * 0.7
	}
	for i, member := range members {
		share := member.Share
		members[i].MemberShareAmt = int(math.Ceil(total_share_amt * float64(share) / 100))
		total_share += share
	}

	application.LeaderShare = 100 - total_share
	application.LeaderShareAmt = int(math.Ceil(total_share_amt * float64(application.LeaderShare) / 100))

	application.MembersInfo = members

	data := app.newTemplateData(r)
	data.Member = members
	data.Application = application
	data.User = user

	app.render(w, http.StatusOK, "admin_view_application.tmpl", data)
}
