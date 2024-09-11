package models

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Application struct {
<<<<<<< HEAD
	SporicRefNo    string `form:"sporic_ref_no"`
	FinancialYear  string `form:"financial_year"`
	ActivityType   string `form:"activity_type"`
	Leader           string `form:"leader"`
	EstimatedAmt   int    `form:"estimated_amt"`
	CompanyName    string `form:"company_name"`
	CompanyAddress string `form:"company_address"`
	ContactPerson  string `form:"contact_person"`
	MailID         string `form:"mail_id"`
	Mobile         string `form:"mobile"`
	Status         int    `form:"status"`
=======
	SporicRefNo    string
	FinancialYear  string
	ActivityType   string
	EstimatedAmt   int
	CompanyName    string
	CompanyAddress string
	ContactPerson  string
	MailID         string
	Mobile         string
	Status         int
>>>>>>> 9ebace6 (jasd)
}

type ProjectStatus = int

const (
	ProjectPendingApproval ProjectStatus = iota
	ProjectApproved
	ProjectCompleted
	ProjectRejected
)

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
<<<<<<< HEAD
			&application.SporicRefNo, &application.FinancialYear, &application.ActivityType, &application.Leader,
=======
			&application.SporicRefNo, &application.FinancialYear, &application.ActivityType,
>>>>>>> 9ebace6 (jasd)
			&application.EstimatedAmt, &application.CompanyName, &application.CompanyAddress, &application.ContactPerson,
			&application.MailID, &application.Mobile, &application.Status,
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

	var count int

	row, err := m.Db.Query("select count(*) from applications where FinancialYear= ?", form.FinancialYear)
	if err != nil {
		log.Fatal(err)
	}
	err = row.Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	sporic_ref_no := "cc" + strings.ToLower(form.ActivityType) + form.FinancialYear + strconv.Itoa(count)

	_, err = m.Db.Exec("insert into applications (sporic_ref_no, financial_year, activity_type, leader, estimated_amt, company_name, company_adress, contact_person_name, contact_person_email, contact_person_mobile, project_status) values (?,?,?,?,?,?,?,?,?,?,?,?,?)", sporic_ref_no, form.FinancialYear, form.ActivityType, form.Leader, form.EstimatedAmt, form.CompanyName, form.CompanyAddress, form.ContactPerson, form.MailID, form.Mobile, ProjectPendingApproval )

	if err != nil {
		log.Fatal(err)
	}
}
