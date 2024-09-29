package models

import "database/sql"

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
	return &payment, nil
}
