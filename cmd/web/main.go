package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
5
	_ "github.com/alexedwards/scs/v2"
	_ "github.com/go-sql-driver/mysql"
)

// var templates = template.Must(template.ParseGlob("template/*.html"))
// var session = scs.New()

type App struct {
	infoLog       *log.Logger
	erroLog       *log.Logger
	Db            *sql.DB
	templateCache map[string]*template.Template
}

func main() {
	addr := flag.String("addr", ":8080", "address to host the site")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	db := loadDatabase()
	app := App{infoLog: infoLog, erroLog: errorLog, templateCache: templateCache, Db: db}

	mux := app.routes()

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  mux,
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func loadDatabase() *sql.DB {
	db, err := sql.Open("mysql", "newuser:newpassword@/sporic")
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db
}

func (app App) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.signin)
	return mux
}