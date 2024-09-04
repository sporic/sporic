package main

import (
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func signin(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}
		email := r.Form.Get("email")
		password := r.Form.Get("password")
		
		if email == "" || password == "" {
			http.Error(w, "Email and password are required", http.StatusBadRequest)
			return
		}
	}
}
