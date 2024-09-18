package models

// TODO: IST Timezones

import (
	"database/sql"
	"errors"
	"strconv"
	"time"
)

type Application struct {
	SporicRefNo              string
	Leader                   int
	FinancialYear            int
	ActivityType             ActivityType
	EstimatedAmt             int
	CompanyName              string
	CompanyAddress           string
	ContactPersonName        string
	ContactPersonDesignation string
	ContactPersonEmail       string
	ContactPersonMobile      string
	Status                   int
	Members                  []string
	StartDate                time.Time
	EndDate                  time.Time
	Payments                 []Payment
	Expenditures             []Expenditure
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

type PaymentStatus = int

const (
	PaymentInvoiceRequested PaymentStatus = iota
	PaymentPending
	PaymentApproved
	PaymentRejected
)

type ExpenditureStatus = int

const (
	ExpenditurePendingApproval ExpenditureStatus = iota
	ExpenditureApproved
	ExpenditureRejected
)

type ApplicationModel struct {
	Db *sql.DB
}

func (m *ApplicationModel) FetchAll() ([]Application, error) {
	rows, err := m.Db.Query("select sporic_ref_no, leader, financial_year, activity_type, estimated_amt, company_name, company_adress, contact_person_name, contact_person_email, contact_person_mobile, contact_person_designation, project_status from applications")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var applications []Application
	for rows.Next() {
		var a Application
		err := rows.Scan(&a.SporicRefNo, &a.Leader, &a.FinancialYear, &a.ActivityType, &a.EstimatedAmt, &a.CompanyName, &a.CompanyAddress, &a.ContactPersonName, &a.ContactPersonEmail, &a.ContactPersonMobile, &a.ContactPersonDesignation, &a.Status)
		if err != nil {
			return nil, err
		}
		rows_members, err := m.Db.Query("Select member_name from team where sporic_ref_no= ?", a.SporicRefNo)

		if err != nil {
			return nil, err
		}

		for rows_members.Next() {
			var member string
			err := rows_members.Scan(&member)
			if err != nil {
				return nil, err
			}
			a.Members = append(a.Members, member)
		}

		rows_payments, err := m.Db.Query("Select payment_id, sporic_ref_no, payment_amt, gst_number, pan_number ,payment_date, payment_status from payment where sporic_ref_no= ?", a.SporicRefNo)

		if err != nil {
			return nil, err
		}

		for rows_payments.Next() {
			var p Payment
			err := rows_payments.Scan(&p.Payment_id, &p.Sporic_ref_no, &p.Payment_amt, &p.Gst_number, &p.Pan_number, &p.Payment_date, &p.Payment_status)
			if err != nil {
				return nil, err
			}
			a.Payments = append(a.Payments, p)
		}

		rows_expenditure, err := m.Db.Query("Select expenditure_name, expenditure_amt, expenditure_date, expenditure_status from expenditure where sporic_ref_no= ?", a.SporicRefNo)
		if err != nil {
			return nil, err
		}

		for rows_expenditure.Next() {
			var e Expenditure
			err := rows_expenditure.Scan(&e.Expenditure_name, &e.Expenditure_amt, &e.Expenditure_date, &e.Expenditure_status)
			if err != nil {
				return nil, err
			}
			a.Expenditures = append(a.Expenditures, e)
		}

		applications = append(applications, a)

	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return applications, nil
}

func (m *ApplicationModel) FetchByLeader(leader int) ([]Application, error) {
	rows, err := m.Db.Query("select sporic_ref_no, leader, financial_year, activity_type, estimated_amt, company_name, company_adress, contact_person_name, contact_person_email, contact_person_mobile, contact_person_designation, project_status from applications where leader=?", leader)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var applications []Application
	for rows.Next() {

		var a Application
		err := rows.Scan(&a.SporicRefNo, &a.Leader, &a.FinancialYear, &a.ActivityType, &a.EstimatedAmt, &a.CompanyName, &a.CompanyAddress, &a.ContactPersonName, &a.ContactPersonEmail, &a.ContactPersonMobile, &a.ContactPersonDesignation, &a.Status)
		if err != nil {
			return nil, err
		}
		rows_members, err := m.Db.Query("Select member_name from team where sporic_ref_no= ?", a.SporicRefNo)
		if err != nil {
			return nil, err
		}

		for rows_members.Next() {
			var member string
			err := rows_members.Scan(&member)
			if err != nil {
				return nil, err
			}
			a.Members = append(a.Members, member)
		}
		rows_payments, err := m.Db.Query("Select payment_id, sporic_ref_no, payment_amt, gst_number, pan_number ,payment_date, payment_status from payment where sporic_ref_no= ?", a.SporicRefNo)
		if err != nil {
			return nil, err
		}

		for rows_payments.Next() {
			var p Payment
			err := rows_payments.Scan(&p.Payment_id, &p.Sporic_ref_no, &p.Payment_amt, &p.Gst_number, &p.Pan_number, &p.Payment_date, &p.Payment_status)
			if err != nil {
				return nil, err
			}
			a.Payments = append(a.Payments, p)
		}

		rows_expenditure, err := m.Db.Query("Select expenditure_name, expenditure_amt, expenditure_date, expenditure_status from expenditure where sporic_ref_no= ?", a.SporicRefNo)
		if err != nil {
			return nil, err
		}

		for rows_expenditure.Next() {
			var e Expenditure
			err := rows_expenditure.Scan(&e.Expenditure_name, &e.Expenditure_amt, &e.Expenditure_date, &e.Expenditure_status)
			if err != nil {
				return nil, err
			}
			a.Expenditures = append(a.Expenditures, e)
		}
		applications = append(applications, a)
	}
	return applications, nil
}

func (m *ApplicationModel) FetchByRefNo(ref_no string) (*Application, error) {
	row := m.Db.QueryRow("select sporic_ref_no, leader, financial_year, activity_type, estimated_amt, company_name, company_adress, contact_person_name, contact_person_email, contact_person_mobile, contact_person_designation, project_status from applications where sporic_ref_no=?", ref_no)
	var a Application
	err := row.Scan(&a.SporicRefNo, &a.Leader, &a.FinancialYear, &a.ActivityType, &a.EstimatedAmt, &a.CompanyName, &a.CompanyAddress, &a.ContactPersonName, &a.ContactPersonEmail, &a.ContactPersonMobile, &a.ContactPersonDesignation, &a.Status)
	if err != nil {
		return nil, err
	}

	rows, err := m.Db.Query("Select member_name from team where sporic_ref_no= ?", a.SporicRefNo)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var member string
		err := rows.Scan(&member)
		if err != nil {
			return nil, err
		}
		a.Members = append(a.Members, member)
	}

	rows, err = m.Db.Query("Select payment_id, sporic_ref_no, payment_amt, gst_number, pan_number ,payment_date, payment_status from payment where sporic_ref_no= ?", a.SporicRefNo)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var p Payment
		err := rows.Scan(&p.Payment_id, &p.Sporic_ref_no, &p.Payment_amt, &p.Gst_number, &p.Pan_number, &p.Payment_date, &p.Payment_status)
		if err != nil {
			return nil, err
		}
		a.Payments = append(a.Payments, p)
	}

	rows_expenditure, err := m.Db.Query("Select expenditure_name, expenditure_amt, expenditure_date, expenditure_status from expenditure where sporic_ref_no= ?", a.SporicRefNo)
	if err != nil {
		return nil, err
	}

	for rows_expenditure.Next() {
		var e Expenditure
		err := rows_expenditure.Scan(&e.Expenditure_name, &e.Expenditure_amt, &e.Expenditure_date, &e.Expenditure_status)
		if err != nil {
			return nil, err
		}
		a.Expenditures = append(a.Expenditures, e)
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

	fin_yr := strconv.Itoa(form.FinancialYear)[2:4] + strconv.Itoa(form.FinancialYear + 1)[2:4]

	sporic_ref_no := "CC" + type_code + fin_yr + strconv.Itoa(count+1)

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
		 contact_person_designation,
		 project_status) 
		 values (?,?,?,?,?,?,?,?,?,?,?,?)`,
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
		form.ContactPersonDesignation,
		ProjectPendingApproval)

	for _, member := range form.Members {
		_, err := m.Db.Exec("insert into team (sporic_ref_no, member_name) values (? ?)", sporic_ref_no, member)
		if err != nil {
			return err
		}
	}

	return err
}

type Payment struct {
	Payment_id     int
	Sporic_ref_no  string
	Payment_amt    int
	Gst_number     string
	Pan_number     string
	Payment_date   sql.NullTime
	Payment_status int
}

func (m *ApplicationModel) Insert_invoice_request(payment Payment) error {

	var application Application

	application.Payments = append(application.Payments, payment)
	_, err := m.Db.Exec(`insert into payment 
	(sporic_ref_no, 
	payment_amt, 
	gst_number, 
	pan_number, 
	payment_status) 
	values (?,?,?,?,?)`,
		payment.Sporic_ref_no,
		payment.Payment_amt,
		payment.Gst_number,
		payment.Pan_number,
		payment.Payment_status)
	if err != nil {
		return err
	}
	return nil
}

type Expenditure struct {
	SporicRefNo        string
	Expenditure_name   string
	Expenditure_amt    int
	Expenditure_date   time.Time
	Expenditure_status int
}

func (m *ApplicationModel) Insert_expenditure(expenditure Expenditure) error {

	var application Application

	application.Expenditures = append(application.Expenditures, expenditure)

	_, err := m.Db.Exec(`insert into expenditure
	(sporic_ref_no,
	expenditure_name,
	expenditure_amt, 
	expenditure_date,
	expenditure_status)
	values (?,?,?,?,?)`,
		expenditure.SporicRefNo,
		expenditure.Expenditure_name,
		expenditure.Expenditure_amt,
		expenditure.Expenditure_date,
		expenditure.Expenditure_status)
	if err != nil {
		return err
	}

	return nil
}
