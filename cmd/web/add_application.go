package main

import (
	"net/http"

	"github.com/sporic/sporic/internal/models"
)

type ApplcationForm struct {
	SporicRefNo    string `form:"sporic_ref_no"`
	FinancialYear  string `form:"financial_year"`
	ActivityType   string `form:"activity_type"`
	Lead           string `form:"lead"`
	EstimatedAmt   int    `form:"estimated_amt"`
	CompanyName    string `form:"company_name"`
	CompanyAddress string `form:"company_address"`
	ContactPerson  string `form:"contact_person"`
	MailID         string `form:"mail_id"`
	Mobile         string `form:"mobile"`
	GST            string `form:"gst"`
	PanNumber      string `form:"pan_number"`
	Status         int    `form:"status"`
}

func (app *App) add_application(w http.ResponseWriter, r *http.Request) {
	var form ApplcationForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusUnprocessableEntity)
		return
	}

	application := models.Application{
		SporicRefNo:    form.SporicRefNo,
		FinancialYear:  form.FinancialYear,
		ActivityType:   form.ActivityType,
		Lead:           form.Lead,
		EstimatedAmt:   form.EstimatedAmt,
		CompanyName:    form.CompanyName,
		CompanyAddress: form.CompanyAddress,
		ContactPerson:  form.ContactPerson,
		MailID:         form.MailID,
		Mobile:         form.Mobile,
		GST:            form.GST,
		PanNumber:      form.PanNumber,
		Status:         form.Status,
	}
	app.applications.Insert(application)
}
