package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/alexedwards/scs/v2"
	_ "github.com/go-sql-driver/mysql"
)

// var templates = template.Must(template.ParseGlob("template/*.html"))
var session = scs.New()

type App struct {
	Db *sql.DB
}

func init() {
	db, err := sql.Open("mysql", "newuser:newpassword@/sporic")
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
}

func main() {
	host := flag.String("host", "", "address to host the site")
	port := flag.Int("port", 8080, "port to host the site")
	flag.Parse()

	app := App{}

	mux := http.NewServeMux()

	mux.HandleFunc("/", app.signin)

	err := http.ListenAndServe(*host+":"+strconv.Itoa(*port), session.LoadAndSave(mux))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
