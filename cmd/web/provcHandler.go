package main

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/sporic/sporic/internal/models"
)

func (app *App) provc_home(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	if user.IsAnonymous() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if user.Role != models.Provc {
		app.notFound(w)
		return
	}

	var applications []models.Application

	applications, err := app.applications.FetchAll()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Applications = applications
	app.render(w, http.StatusOK, "provc_home.tmpl", data)
}

func (app *App) provc_view_application(w http.ResponseWriter, r *http.Request) {

	user := app.contextGetUser(r)
	if user.IsAnonymous() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if user.Role != models.Provc {
		app.notFound(w)
		return
	}
	params := httprouter.ParamsFromContext(r.Context())

	refno := params.ByName("refno")

	application, err := app.applications.FetchByRefNo(refno)

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		action := r.Form.Get("action")
		if action == "approve" {
			err = app.applications.SetStatus(refno, models.ProjectApprovedByProVC)
			if err != nil {
				app.serverError(w, err)
				return
			}

			var notification models.Notification

			notification.CreatedAt = time.Now()
			notification.NotiType = models.ProVCApproved
			notification.Description = fmt.Sprintf(models.NotificationTypeMap[models.ProVCApproved], refno)
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
			faculty := strconv.Itoa(application.Leader)
			recievers := append(accounts, admins...)
			recievers = append(recievers, faculty)
			notification.To = recievers
			err = app.notifications.SendNotification(notification, app.mailer)
			if err != nil {
				app.serverError(w, err)
				return
			}

			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
	}

	fmt.Printf("refno: %s\n", refno)
	if err != nil {
		app.serverError(w, err)
		return
	}

	var TotalAmt int = 0
	var TotalTax int = 0
	for _, payment := range application.Payments {
		TotalAmt += payment.Payment_amt
		TotalTax += payment.Tax * payment.Payment_amt / 100
	}
	application.TotalAmount = TotalAmt
	application.Taxes = TotalTax

	application.TotalAmountIncludeTax = TotalAmt + TotalTax

	var TotalExpenditure int = 0
	for _, expenditure := range application.Expenditures {
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
	app.render(w, http.StatusOK, "provc_view_application.tmpl", data)
}
