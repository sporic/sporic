package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"

	"github.com/go-playground/form"
	"github.com/sporic/sporic/internal/models"
	"github.com/xuri/excelize/v2"
)

func (app *App) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *App) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *App) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *App) decodePostForm(_, dst any, values url.Values) error {

	err := app.formDecoder.Decode(dst, values)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError
		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}
		return err
	}
	return nil
}

func (app *App) handleFile(r *http.Request, folder_name string, prefix string, file_type FileType, field_name string) error {
	file, _, err := r.FormFile(field_name)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = app.applications.FetchByRefNo(folder_name)
	if err != nil {
		return err
	}

	uploadDir := "documents/" + folder_name
	err = os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		return err
	}

	filename := ""
	if file_type == ProposalDoc {

		_, err = app.applications.FetchByRefNo(prefix)
		if err != nil {
			return nil
		}

		filename = folder_name + "_" + prefix + "_proposal" + ".pdf"
	}
	if file_type == Invoice {

		prefix, _ := strconv.Atoi(prefix)
		payment, err := app.payments.GetPaymentById(prefix)
		if err != nil {
			return nil
		}

		if payment.Payment_id != prefix {
			return nil
		}

		filename = folder_name + "_" + strconv.Itoa(prefix) + "_invoice" + ".pdf"
	}
	if file_type == PaymentProof {

		prefix, _ := strconv.Atoi(prefix)
		payment, err := app.payments.GetPaymentById(prefix)
		if err != nil {
			return nil
		}

		if payment.Payment_id != prefix {
			return nil
		}

		filename = folder_name + "_" + strconv.Itoa(prefix) + "_payment" + ".pdf"
	}
	if file_type == GstCirtificate {

		prefix, _ := strconv.Atoi(prefix)
		_, err := app.payments.GetPaymentById(prefix)
		if err != nil {
			return err
		}
		filename = folder_name + "_" + strconv.Itoa(prefix) + "_tax_cirtificate" + ".pdf"
	}
	if file_type == PanCard {
		prefix, _ := strconv.Atoi(prefix)
		payment, err := app.payments.GetPaymentById(prefix)
		if err != nil {
			return nil
		}

		if payment.Payment_id != prefix {
			return nil
		}

		filename = folder_name + "_" + strconv.Itoa(prefix) + "_tax_cirtificate" + ".pdf"
	}
	if file_type == CompletionDoc {

		_, err = app.applications.FetchByRefNo(prefix)
		if err != nil {
			return nil
		}

		filename = folder_name + "_" + prefix + "_completion_form" + ".pdf"
	}
	if file_type == ExpenditureProof {

		prefix, err := strconv.Atoi(prefix)
		if err != nil {
			return err
		}
		expenditure, err := app.applications.GetExpenditureById(prefix)
		if err != nil {
			return err
		}

		if expenditure.Expenditure_id != prefix {
			return nil
		}

		filename = folder_name + "_" + strconv.Itoa(prefix) + "_expenditure_proof" + ".pdf"
	}
	if file_type == ExpenditureInvoice {

		prefix, _ := strconv.Atoi(prefix)
		expenditure, err := app.applications.GetExpenditureById(prefix)
		if err != nil {
			return err
		}

		if expenditure.Expenditure_id != prefix {
			return nil
		}
		filename = folder_name + "_" + strconv.Itoa(prefix) + "_expenditure_invoice" + ".pdf"
	}
	if file_type == FeedbackForm {

		_, err = app.applications.FetchByRefNo(prefix)
		if err != nil {
			return nil
		}

		filename = folder_name + "_" + prefix + "_feedback_form" + ".pdf"
	}

	filePath := filepath.Join(uploadDir, filename)

	destFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, file)
	if err != nil {
		return err
	}
	return nil
}

func (app *App) GenerateExcel(applications []models.Application) (*excelize.File, error) {

	f := excelize.NewFile()
	index, err := f.NewSheet("Sheet1")

	if err != nil {
		app.errorLog.Println(err)
		return nil, err
	}

	headers := []string{
		"Sporic ID",
		"Project Title",
		"Leader",
		"Financial Year",
		"Activity Type",
		"Estimated Amount",
		"Total Payment",
		"Total Tax",
		"Total Amount Paid",
		"Total Expenditure",
		"Company Name",
		"Company Address",
		"Billing Address",
		"Contact Person Name",
		"Contact Person Designation",
		"Contact Person Email",
		"Contact Person Mobile",
		"Status",
		"Members",
		"Member Students",
		"Start Date",
		"End Date",
		"Comments",
		"Completion Date",
		"Resource Used",
	}

	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue("Sheet1", cell, header)
	}

	for i, row := range applications {
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), row.SporicRefNo)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), row.ProjectTitle)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), row.Leader)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+2), row.FinancialYear)
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", i+2), row.ActivityType)
		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", i+2), row.EstimatedAmt)
		TotalPayment := 0
		TotalTax := 0
		for _, payment := range row.Payments {
			if payment.Payment_status == models.PaymentCompleted {
				TotalPayment += payment.Payment_amt
				TotalTax += payment.Tax
			}
		}
		f.SetCellValue("Sheet1", fmt.Sprintf("G%d", i+2), TotalPayment)
		f.SetCellValue("Sheet1", fmt.Sprintf("H%d", i+2), TotalTax)
		f.SetCellValue("Sheet1", fmt.Sprintf("I%d", i+2), TotalPayment+TotalTax)
		TotalExpenditure := 0
		for _, expenditure := range row.Expenditures {
			if expenditure.Expenditure_status == models.ExpenditureApproved {
				TotalExpenditure += expenditure.Expenditure_amt
			}
		}
		f.SetCellValue("Sheet1", fmt.Sprintf("J%d", i+2), TotalExpenditure)
		f.SetCellValue("Sheet1", fmt.Sprintf("K%d", i+2), row.CompanyName)
		f.SetCellValue("Sheet1", fmt.Sprintf("L%d", i+2), row.CompanyAddress)
		f.SetCellValue("Sheet1", fmt.Sprintf("M%d", i+2), row.BillingAddress)
		f.SetCellValue("Sheet1", fmt.Sprintf("N%d", i+2), row.ContactPersonName)
		f.SetCellValue("Sheet1", fmt.Sprintf("O%d", i+2), row.ContactPersonDesignation)
		f.SetCellValue("Sheet1", fmt.Sprintf("P%d", i+2), row.ContactPersonEmail)
		f.SetCellValue("Sheet1", fmt.Sprintf("Q%d", i+2), row.ContactPersonMobile)
		f.SetCellValue("Sheet1", fmt.Sprintf("R%d", i+2), row.Status)

		members := ""
		for _, member := range row.Members {
			members += ", " + member
		}
		f.SetCellValue("Sheet1", fmt.Sprintf("S%d", i+2), members)

		memberStudents := ""
		for _, memberStudent := range row.MemberStudents {
			memberStudents += ", " + memberStudent
		}
		f.SetCellValue("Sheet1", fmt.Sprintf("T%d", i+2), memberStudents)

		f.SetCellValue("Sheet1", fmt.Sprintf("U%d", i+2), row.StartDate)
		f.SetCellValue("Sheet1", fmt.Sprintf("V%d", i+2), row.EndDate)
		f.SetCellValue("Sheet1", fmt.Sprintf("W%d", i+2), row.Comments)
		f.SetCellValue("Sheet1", fmt.Sprintf("X%d", i+2), row.CompletionDate)
		f.SetCellValue("Sheet1", fmt.Sprintf("Y%d", i+2), row.ResourceUsed)
	}

	f.SetActiveSheet(index)

	return f, nil
}
