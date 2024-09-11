package models

import (
	"database/sql"
	"fmt"
	"log"
)

type Application struct {
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

type ApplicationModel struct {
	Db *sql.DB
}

func (m *ApplicationModel) Fetch(sporic_ref_no string, leader string) []Application {

	var applications []Application

	var rows *sql.Rows
	var err error
	if sporic_ref_no == "" {
		rows, err = m.Db.Query("SELECT * FROM applications WHERE leader = ?", leader)
	} else if leader == "" {
		rows, err = m.Db.Query("SELECT * FROM applications WHERE sporic_ref_no = ?", sporic_ref_no)
	} else {
		rows, err = m.Db.Query("SELECT * FROM applications")
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

func (m *ApplicationModel) Insert(form Application) {

}
