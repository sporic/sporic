package models

// TODO: IST Timezones

import (
	"time"
)

type ExpenditureStatus = int

// 0->1/2->3
const (
	ExpenditurePendingApproval ExpenditureStatus = iota
	ExpenditureApproved
	ExpenditureRejected
	ExpenditureCompleted
)

func (m *ApplicationModel) SetExpenditureStatus(exp_id string, status ExpenditureStatus) error {
	_, err := m.Db.Exec("update expenditure set expenditure_status = ? where expenditure_id = ?", status, exp_id)
	if err != nil {
		return err
	}
	return nil
}

type Expenditure struct {
	Expenditure_type   int
	Expenditure_id     int
	SporicRefNo        string
	Expenditure_name   string
	Expenditure_amt    int
	Expenditure_date   time.Time
	Expenditure_status ExpenditureStatus
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

func (m *ApplicationModel) GetExpenditureByRefNo(sporic_ref_no string) ([]Expenditure, error) {

	rows, err := m.Db.Query("select expenditure_id,expenditure_name, expenditure_amt, expenditure_date, expenditure_status, expenditure_type from expenditure where sporic_ref_no =?", sporic_ref_no)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenditure []Expenditure
	for rows.Next() {
		var e Expenditure
		err := rows.Scan(&e.Expenditure_id, &e.Expenditure_name, &e.Expenditure_amt, &e.Expenditure_date, &e.Expenditure_status, &e.Expenditure_type)
		if err != nil {
			return nil, err
		}

		expenditure = append(expenditure, e)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return expenditure, nil
}

func (m *ApplicationModel) GetExpenditureById(expenditure_id int) (*Expenditure, error) {

	row := m.Db.QueryRow("select sporic_ref_no,expenditure_name,expenditure_amt,expenditure_date,expenditure_status,expenditure_type from expenditure where expenditure_id =?", expenditure_id)

	var expenditure Expenditure

	err := row.Scan(&expenditure.SporicRefNo, &expenditure.Expenditure_name, &expenditure.Expenditure_amt, &expenditure.Expenditure_date, &expenditure.Expenditure_status, &expenditure.Expenditure_type)

	if err != nil {
		return nil, err
	}

	return &expenditure, nil
}
