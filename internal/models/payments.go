package models

import (
	"database/sql"
	"time"
)

type PaymentStatus = int

// 0->4->1->5->2/3
const (
	PaymentInvoiceRequested PaymentStatus = iota
	PaymentPending
	PaymentCompleted
	PaymentRejected
	PaymentInvoiceForwarded
	PaymentProofUploaded
)

type Payment struct {
	Payment_id     int
	Currency       string
	Transaction_id string
	Sporic_ref_no  string
	Payment_amt    int
	Tax            int
	Total_amt      int
	Gst_number     string
	Pan_number     string
	Payment_date   sql.NullTime
	Payment_status int
}

type PaymentModel struct {
	Db *sql.DB
}

func (p *PaymentModel) GetAllPayments() ([]Payment, error) {

	rows, err := p.Db.Query("select payment_id,currency, tax,  sporic_ref_no, payment_amt, gst_number, pan_number, payment_date, payment_status, transaction_id from payment")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []Payment
	for rows.Next() {
		var p Payment
		err := rows.Scan(&p.Payment_id, &p.Currency, &p.Tax, &p.Sporic_ref_no, &p.Payment_amt, &p.Gst_number, &p.Pan_number, &p.Payment_date, &p.Payment_status, &p.Transaction_id)
		if err != nil {
			return nil, err
		}
		p.Total_amt = p.Payment_amt + p.Tax*p.Payment_amt/100
		payments = append(payments, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return payments, nil
}

func (p *PaymentModel) GetPaymentByRefNo(sporic_ref_no string) ([]Payment, error) {

	rows, err := p.Db.Query("select payment_id,currency, tax, payment_amt, gst_number, pan_number, payment_date, payment_status, transaction_id from payment where sporic_ref_no =?", sporic_ref_no)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []Payment
	for rows.Next() {
		var p Payment
		err := rows.Scan(&p.Payment_id, &p.Currency, &p.Tax, &p.Payment_amt, &p.Gst_number, &p.Pan_number, &p.Payment_date, &p.Payment_status, &p.Transaction_id)
		if err != nil {
			return nil, err
		}
		p.Total_amt = p.Payment_amt + p.Tax*p.Payment_amt/100
		payments = append(payments, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return payments, nil
}

func (p *PaymentModel) GetPaymentById(payment_id int) (*Payment, error) {

	row := p.Db.QueryRow("select payment_id, sporic_ref_no, payment_amt, gst_number, pan_number, payment_date, payment_status, transaction_id, currency, tax from payment where payment_id = ?", payment_id)

	var payment Payment
	err := row.Scan(&payment.Payment_id, &payment.Sporic_ref_no, &payment.Payment_amt, &payment.Gst_number, &payment.Pan_number, &payment.Payment_date, &payment.Payment_status, &payment.Transaction_id, &payment.Currency, &payment.Tax)
	if err != nil {
		return nil, err
	}
	payment.Total_amt = payment.Payment_amt + payment.Tax*payment.Payment_amt/100
	return &payment, nil
}

func (m *ApplicationModel) SetPaymentStatus(payment_id string, status PaymentStatus) error {
	_, err := m.Db.Exec("update payment set payment_status = ? where payment_id = ?", status, payment_id)
	if err != nil {
		return err
	}
	return nil
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

func (m *ApplicationModel) UpdatePayment(payment Payment) error {
	_, err := m.Db.Exec("update payment set transaction_id = ?, payment_date=?, payment_status=? where payment_id = ?", payment.Transaction_id, time.Now(), PaymentProofUploaded, payment.Payment_id)

	if err != nil {
		return err
	}
	return nil
}
