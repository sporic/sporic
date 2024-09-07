package models

import (
	"database/sql"
	"fmt"
	"log"
)

type Application struct {
	SporicRefNo    string `json:"sporic_ref_no"`
	FinancialYear  string `json:"financial_year"`
	ActivityType   string `json:"activity_type"`
	Lead           string `json:"lead"`
	EstimatedAmt   int    `json:"estimated_amt"`
	CompanyName    string `json:"company_name"`
	CompanyAddress string `json:"company_address"`
	ContactPerson  string `json:"contact_person"`
	MailID         string `json:"mail_id"`
	Mobile         string `json:"mobile"`
	GST            string `json:"gst"`
	PanNumber      string `json:"pan_number"`
	Status         int    `json:"status"`
}

type ApplicationModel struct {
	Db *sql.DB
}

func (m *ApplicationModel) Fetch_applications(sporic_ref_no string, leader string) []Application {

	var applications []Application

	var rows *sql.Rows
	var err error
	if sporic_ref_no == "" {
		rows, err = m.Db.Query("SELECT * FROM applications WHERE leader = ?", leader)
	} else {
		rows, err = m.Db.Query("SELECT * FROM applications WHERE sporic_ref_no = ?", sporic_ref_no)
	}

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var application Application
		err := rows.Scan(
			&application.SporicRefNo, &application.FinancialYear, &application.ActivityType, &application.Lead,
			&application.EstimatedAmt, &application.CompanyName, &application.CompanyAddress, &application.ContactPerson,
			&application.MailID, &application.Mobile, &application.GST, &application.PanNumber, &application.Status,
		)
		if err != nil {
			log.Fatal(err)
		}
		applications = append(applications, application)

	}
	fmt.Println(applications)

	return applications
}
