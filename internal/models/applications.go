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
	ProjectTitle             string
	Leader                   int
	FinancialYear            int
	ActivityType             ActivityType
	EstimatedAmt             int
	CompanyName              string
	CompanyAddress           string
	BillingAddress           string
	ContactPersonName        string
	ContactPersonDesignation string
	ContactPersonEmail       string
	ContactPersonMobile      string
	Status                   int
	Members                  []string
	MemberStudents           []string
	StartDate                time.Time
	EndDate                  time.Time
	Payments                 []Payment
	Expenditures             []Expenditure
	Comments                 string
	CompletionDate           time.Time
	ResourceUsed             int
}

type ProjectStatus = int

const (
	ProjectPendingApproval ProjectStatus = iota
	ProjectApproved
	ProjectCompleted
	ProjectCompleteApprovalPending
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
	PaymentInvoiceForwarded
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
	rows, err := m.Db.Query("select sporic_ref_no, project_title, leader, financial_year, activity_type, estimated_amt, company_name, company_adress, billing_address,  contact_person_name, contact_person_email, contact_person_mobile, contact_person_designation, project_status, comments, resources_used, completion_date from applications")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var applications []Application
	for rows.Next() {
		var a Application
		err := rows.Scan(&a.SporicRefNo, &a.ProjectTitle, &a.Leader, &a.FinancialYear, &a.ActivityType, &a.EstimatedAmt, &a.CompanyName, &a.CompanyAddress, &a.BillingAddress, &a.ContactPersonName, &a.ContactPersonEmail, &a.ContactPersonMobile, &a.ContactPersonDesignation, &a.Status, &a.Comments, &a.ResourceUsed, &a.CompletionDate)
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

		rows_member_student, err := m.Db.Query("Select member_name from team_student where sporic_ref_no= ?", a.SporicRefNo)

		if err != nil {
			return nil, err
		}

		for rows_member_student.Next() {
			var member string
			err := rows_member_student.Scan(&member)
			if err != nil {
				return nil, err
			}
			a.MemberStudents = append(a.MemberStudents, member)
		}

		rows_payments, err := m.Db.Query("Select payment_id, sporic_ref_no, currency, payment_amt,tax, gst_number, pan_number ,payment_date, payment_status, transaction_id from payment where sporic_ref_no= ?", a.SporicRefNo)

		if err != nil {
			return nil, err
		}

		for rows_payments.Next() {
			var p Payment
			err := rows_payments.Scan(&p.Payment_id, &p.Sporic_ref_no, &p.Currency, &p.Payment_amt, &p.Tax, &p.Gst_number, &p.Pan_number, &p.Payment_date, &p.Payment_status, &p.Transaction_id)
			if err != nil {
				return nil, err
			}
			a.Payments = append(a.Payments, p)
		}

		rows_expenditure, err := m.Db.Query("Select expenditure_id, expenditure_type, sporic_ref_no, expenditure_name, expenditure_amt, expenditure_date, expenditure_status from expenditure where sporic_ref_no= ?", a.SporicRefNo)
		if err != nil {
			return nil, err
		}

		for rows_expenditure.Next() {
			var e Expenditure
			err := rows_expenditure.Scan(&e.Expenditure_id, &e.Expenditure_type, &e.SporicRefNo, &e.Expenditure_name, &e.Expenditure_amt, &e.Expenditure_date, &e.Expenditure_status)
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
	rows, err := m.Db.Query("select sporic_ref_no, project_title, leader, financial_year, activity_type, estimated_amt, company_name, company_adress, billing_address, contact_person_name, contact_person_email, contact_person_mobile, contact_person_designation, project_status, comments, resources_used, completion_date from applications where leader=?", leader)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var applications []Application
	for rows.Next() {

		var a Application
		err := rows.Scan(&a.SporicRefNo, &a.ProjectTitle, &a.Leader, &a.FinancialYear, &a.ActivityType, &a.EstimatedAmt, &a.CompanyName, &a.CompanyAddress, &a.BillingAddress, &a.ContactPersonName, &a.ContactPersonEmail, &a.ContactPersonMobile, &a.ContactPersonDesignation, &a.Status, &a.Comments, &a.ResourceUsed, &a.CompletionDate)
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

		rows_member_student, err := m.Db.Query("Select member_name from team_student where sporic_ref_no= ?", a.SporicRefNo)

		if err != nil {
			return nil, err
		}

		for rows_member_student.Next() {
			var member string
			err := rows_member_student.Scan(&member)
			if err != nil {
				return nil, err
			}
			a.MemberStudents = append(a.MemberStudents, member)
		}

		rows_payments, err := m.Db.Query("Select payment_id, sporic_ref_no,currency, payment_amt,tax, gst_number, pan_number ,payment_date, payment_status, transaction_id from payment where sporic_ref_no= ?", a.SporicRefNo)
		if err != nil {
			return nil, err
		}

		for rows_payments.Next() {
			var p Payment
			err := rows_payments.Scan(&p.Payment_id, &p.Sporic_ref_no, &p.Currency, &p.Payment_amt, &p.Tax, &p.Gst_number, &p.Pan_number, &p.Payment_date, &p.Payment_status, &p.Transaction_id)
			if err != nil {
				return nil, err
			}
			a.Payments = append(a.Payments, p)
		}

		rows_expenditure, err := m.Db.Query("Select expenditure_id,expenditure_type, sporic_ref_no, expenditure_name, expenditure_amt, expenditure_date, expenditure_status from expenditure where sporic_ref_no= ?", a.SporicRefNo)
		if err != nil {
			return nil, err
		}

		for rows_expenditure.Next() {
			var e Expenditure
			err := rows_expenditure.Scan(&e.Expenditure_id, &e.Expenditure_type, &e.SporicRefNo, &e.Expenditure_name, &e.Expenditure_amt, &e.Expenditure_date, &e.Expenditure_status)
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
	row := m.Db.QueryRow("select sporic_ref_no, project_title, leader, financial_year, activity_type, estimated_amt, company_name, company_adress, billing_address, contact_person_name, contact_person_email, contact_person_mobile, contact_person_designation, project_status,comments, resources_used, completion_date from applications where sporic_ref_no=?", ref_no)
	var a Application
	err := row.Scan(&a.SporicRefNo, &a.ProjectTitle, &a.Leader, &a.FinancialYear, &a.ActivityType, &a.EstimatedAmt, &a.CompanyName, &a.CompanyAddress, &a.BillingAddress, &a.ContactPersonName, &a.ContactPersonEmail, &a.ContactPersonMobile, &a.ContactPersonDesignation, &a.Status, &a.Comments, &a.ResourceUsed, &a.CompletionDate)
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

	rows_member_student, err := m.Db.Query("Select member_name from team_student where sporic_ref_no= ?", a.SporicRefNo)

	if err != nil {
		return nil, err
	}

	for rows_member_student.Next() {
		var member string
		err := rows_member_student.Scan(&member)
		if err != nil {
			return nil, err
		}
		a.MemberStudents = append(a.MemberStudents, member)
	}

	rows, err = m.Db.Query("Select payment_id, sporic_ref_no, currency, payment_amt,tax, gst_number, pan_number ,payment_date, payment_status, transaction_id from payment where sporic_ref_no= ?", a.SporicRefNo)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var p Payment
		err := rows.Scan(&p.Payment_id, &p.Sporic_ref_no, &p.Currency, &p.Payment_amt, &p.Tax, &p.Gst_number, &p.Pan_number, &p.Payment_date, &p.Payment_status, &p.Transaction_id)
		if err != nil {
			return nil, err
		}
		a.Payments = append(a.Payments, p)
	}

	rows_expenditure, err := m.Db.Query("Select expenditure_id, expenditure_type, sporic_ref_no, expenditure_name, expenditure_amt, expenditure_date, expenditure_status from expenditure where sporic_ref_no= ?", a.SporicRefNo)
	if err != nil {
		return nil, err
	}

	for rows_expenditure.Next() {
		var e Expenditure
		err := rows_expenditure.Scan(&e.Expenditure_id, &e.Expenditure_type, &e.SporicRefNo, &e.Expenditure_name, &e.Expenditure_amt, &e.Expenditure_date, &e.Expenditure_status)
		if err != nil {
			return nil, err
		}
		a.Expenditures = append(a.Expenditures, e)
	}

	return &a, nil
}

func (m *ApplicationModel) Insert(form Application) (string, error) {

	var count int

	row := m.Db.QueryRow("select count(*) from applications where financial_year=?", form.FinancialYear)
	err := row.Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		count = 0
	} else if err != nil {
		return "", err
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
		 project_title,
		 leader, 
		 financial_year, 
		 activity_type, 
		 estimated_amt, 
		 company_name, 
		 company_adress,
		 billing_address, 
		 contact_person_name, 
		 contact_person_email, 
		 contact_person_mobile,
		 contact_person_designation,
		 project_status) 
		 values (?,?,?,?,?,?,?,?,?,?,?,?, ?, ?)`,
		sporic_ref_no,
		form.ProjectTitle,
		form.Leader,
		form.FinancialYear,
		form.ActivityType,
		form.EstimatedAmt,
		form.CompanyName,
		form.CompanyAddress,
		form.BillingAddress,
		form.ContactPersonName,
		form.ContactPersonEmail,
		form.ContactPersonMobile,
		form.ContactPersonDesignation,
		ProjectPendingApproval)
	if err != nil {
		return "", err
	}

	for _, member := range form.Members {
		_, err := m.Db.Exec("insert into team (sporic_ref_no, member_name) values (?, ?)", sporic_ref_no, member)
		if err != nil {
			return "", err
		}
	}

	for _, member := range form.MemberStudents {
		_, err := m.Db.Exec("insert into team_student (sporic_ref_no, member_name) values (?, ?)", sporic_ref_no, member)
		if err != nil {
			return "", err
		}
	}

	return sporic_ref_no, err
}

func (m *ApplicationModel) SetStatus(refno string, status ProjectStatus) error {
	_, err := m.Db.Exec("update applications set project_status = ? where sporic_ref_no = ?", status, refno)
	if err != nil {
		return err
	}
	return nil
}

func (m *ApplicationModel) SetExpenditureStatus(refno string, status ExpenditureStatus) error {
	_, err := m.Db.Exec("update expenditure set expenditure_status = ? where sporic_ref_no = ?", status, refno)
	if err != nil {
		return err
	}
	return nil
}

func (m *ApplicationModel) SetPaymentStatus(refno string, status PaymentStatus) error {
	_, err := m.Db.Exec("update payment set payment_status = ? where sporic_ref_no = ?", status, refno)
	if err != nil {
		return err
	}

	return nil
}

type Payment struct {
	Payment_id     int
	Currency       string
	Transaction_id string
	Sporic_ref_no  string
	Payment_amt    int
	Tax            int
	Gst_number     string
	Pan_number     string
	Payment_date   sql.NullTime
	Payment_status int
}

func (m *ApplicationModel) Insert_invoice_request(payment Payment) (int, error) {

	var application Application

	application.Payments = append(application.Payments, payment)
	res, err := m.Db.Exec(`insert into payment 
	(sporic_ref_no,
	currency, 
	payment_amt,
	tax, 
	gst_number, 
	pan_number, 
	payment_status) 
	values (?,?,?,?,?,?,?)`,
		payment.Sporic_ref_no,
		payment.Currency,
		payment.Payment_amt,
		payment.Tax,
		payment.Gst_number,
		payment.Pan_number,
		payment.Payment_status)
	if err != nil {
		return -1, err
	}

	lastInsertId, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	return int(lastInsertId), err
}

type Expenditure struct {
	Expenditure_type   int
	Expenditure_id     int
	SporicRefNo        string
	Expenditure_name   string
	Expenditure_amt    int
	Expenditure_date   time.Time
	Expenditure_status int
}

func (m *ApplicationModel) Insert_expenditure(expenditure Expenditure) (int, error) {

	var application Application

	application.Expenditures = append(application.Expenditures, expenditure)

	res, err := m.Db.Exec(`insert into expenditure
	(sporic_ref_no,
	expenditure_type,
	expenditure_name,
	expenditure_amt, 
	expenditure_date,
	expenditure_status)
	values (?,?,?,?,?,?)`,
		expenditure.SporicRefNo,
		expenditure.Expenditure_type,
		expenditure.Expenditure_name,
		expenditure.Expenditure_amt,
		expenditure.Expenditure_date,
		expenditure.Expenditure_status)
	if err != nil {
		return -1, err
	}

	lastInsertId, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return int(lastInsertId), nil
}

type Completion struct {
	SporicRefNo    string
	Comments       string
	CompletionDate time.Time
	ResourceUsed   int
}

func (m *ApplicationModel) Complete_Project(completion Completion) error {

	_, err := m.Db.Exec("update applications set project_status = ?,resources_used = ?, comments = ?  where sporic_ref_no = ?", ProjectCompleteApprovalPending, completion.ResourceUsed, completion.Comments, completion.SporicRefNo)

	if err != nil {
		return err
	}
	return nil
}

func (m *ApplicationModel) UpdatePayment(payment Payment) error {
	_, err := m.Db.Exec("update payment set transaction_id = ? where payment_id = ?", payment.Transaction_id, payment.Payment_id)

	if err != nil {
		return err
	}
	return nil
}
