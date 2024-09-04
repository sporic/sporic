package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

type templateData struct{}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		files := []string{
			"./ui/html/base.tmpl", "./ui/html/partials/nav.tmpl", page,
		}
		ts, err := template.ParseFiles(files...)
		if err != nil {
			return nil, err
		}
		cache[name] = ts
	}
	return cache, nil
}

func (app *App) render(w http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, err)
		return
	}
	// Write out the provided HTTP status code ('200 OK', '400 Bad Request' // etc).
	w.WriteHeader(status)
	// Execute the template set and write the response body. Again, if there // is any error we call the the serverError() helper.
	err := ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, err)
	}
}
