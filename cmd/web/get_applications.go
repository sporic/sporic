package main

import (
	"fmt"
	"net/http"
)

func (app App) get_applications(w http.ResponseWriter, r *http.Request) {

	sporic_ref_no := r.Form.Get("sporic_ref_no")
	leader := r.Form.Get("leader")
	// to do validate params
	applications := app.applications.Fetch_applications(sporic_ref_no, leader)
	fmt.Println(applications)
}
