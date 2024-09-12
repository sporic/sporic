package models

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
)

type Application struct {
	SporicRefNo         string
	Leader              int
	FinancialYear       int
	ActivityType        ActivityType
	EstimatedAmt        int
	CompanyName         string
	CompanyAddress      string
	ContactPersonName   string
	ContactPersonEmail  string
	ContactPersonMobile string
	Status              int
	Members             []string
}

type ProjectStatus = int

const (
	ProjectPendingApproval ProjectStatus = iota
	ProjectApproved
	ProjectCompleted
	ProjectRejected
)

type ActivityType = int

const (
	ActivityTypeConsultancy ActivityType = iota
	ActivityTypeTraining
)

type ApplicationModel struct {
	Db *sql.DB
}

func (m *ApplicationModel) FetchAll() ([]Application, error) {
	rows, err := m.Db.Query("select sporic_ref_no, leader, financial_year, activity_type, estimated_amt, company_name, company_adress, contact_person_name, contact_person_email, contact_person_mobile, project_status from applications")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var applications []Application
	for rows.Next() {
		var a Application
		err := rows.Scan(&a.SporicRefNo, &a.Leader, &a.FinancialYear, &a.ActivityType, &a.EstimatedAmt, &a.CompanyName, &a.CompanyAddress, &a.ContactPersonName, &a.ContactPersonEmail, &a.ContactPersonMobile, &a.Status)
		if err != nil {
			return nil, err
		}
		applications = append(applications, a)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return applications, nil
}

func (m *ApplicationModel) FetchByLeader(leader int) ([]Application, error) {
	rows, err := m.Db.Query("select sporic_ref_no, leader, financial_year, activity_type, estimated_amt, company_name, company_adress, contact_person_name, contact_person_email, contact_person_mobile, project_status from applications where leader=?", leader)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var applications []Application
	for rows.Next() {
		var a Application
		err := rows.Scan(&a.SporicRefNo, &a.Leader, &a.FinancialYear, &a.ActivityType, &a.EstimatedAmt, &a.CompanyName, &a.CompanyAddress, &a.ContactPersonName, &a.ContactPersonEmail, &a.ContactPersonMobile, &a.Status)
		if err != nil {
			return nil, err
		}
		applications = append(applications, a)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return applications, nil
}

func (m *ApplicationModel) FetchByRefNo(ref_no string) (*Application, error) {
	row := m.Db.QueryRow("select sporic_ref_no, leader, financial_year, activity_type, estimated_amt, company_name, company_adress, contact_person_name, contact_person_email, contact_person_mobile, project_status from applications where sporic_ref_no=?", ref_no)
	var a Application
	err := row.Scan(&a.SporicRefNo, &a.Leader, &a.FinancialYear, &a.ActivityType, &a.EstimatedAmt, &a.CompanyName, &a.CompanyAddress, &a.ContactPersonName, &a.ContactPersonEmail, &a.ContactPersonMobile, &a.Status)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (m *ApplicationModel) Insert(form Application) error {

	var count int

	row := m.Db.QueryRow("select count(*) from applications where financial_year=?", form.FinancialYear)
	err := row.Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		count = 0
	} else if err != nil {
		return err
	}

	type_code := ""
	switch form.ActivityType {
	case ActivityTypeConsultancy:
		type_code = "CP"
	case ActivityTypeTraining:
		type_code = "IT"
	}

	sporic_ref_no := "CC" + type_code + strconv.Itoa(form.FinancialYear) + strconv.Itoa(count+1)

	_, err = m.Db.Exec(`insert into applications 
		(sporic_ref_no,
		 leader, 
		 financial_year, 
		 activity_type, 
		 estimated_amt, 
		 company_name, 
		 company_adress, 
		 contact_person_name, 
		 contact_person_email, 
		 contact_person_mobile,
		 project_status) 
		 values (?,?,?,?,?,?,?,?,?,?,?)`,
		sporic_ref_no,
		form.Leader,
		form.FinancialYear,
		form.ActivityType,
		form.EstimatedAmt,
		form.CompanyName,
		form.CompanyAddress,
		form.ContactPersonName,
		form.ContactPersonEmail,
		form.ContactPersonMobile,
		ProjectPendingApproval)

	for _, member := range form.Members {
		_, err := m.Db.Exec("insert into team (sporic_ref_no, member) values (? ?)", sporic_ref_no, member)
		if err != nil {
			log.Println(err)
		}
	}

	return err
}
