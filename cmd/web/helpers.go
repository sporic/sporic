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

	"github.com/go-playground/form"
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

	uploadDir := "documents/" + folder_name
	err = os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		return err
	}

	filename := ""
	if file_type == ProposalDoc {
		filename = prefix + "_proposal" + ".pdf"
	}
	if file_type == Invoice {
		filename = prefix + "_invoice" + ".pdf"
	}
	if file_type == PaymentProof {
		filename = prefix + "_payment" + ".pdf"
	}
	if file_type == GstCirtificate {
		filename = prefix + "_gst_cirtificate" + ".pdf"
	}
	if file_type == PanCard {
		filename = prefix + "_pan_card" + ".pdf"
	}
	if file_type == CompletionDoc {
		filename = prefix + "_completion_form" + ".pdf"
	}
	if file_type == ExpenditureProof {
		filename = prefix + "_expenditure_proof" + ".pdf"
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
