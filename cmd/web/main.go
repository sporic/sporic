package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
)

// var templates = template.Must(template.ParseGlob("template/*.html"))
// var session = scs.New()

type App struct {
	infoLog       *log.Logger
	errorLog      *log.Logger
	db            *sql.DB
	templateCache map[string]*template.Template
	formDecoder   *form.Decoder
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

	formDecoder := form.NewDecoder()

	app := App{
		infoLog:       infoLog,
		errorLog:      errorLog,
		templateCache: templateCache,
		db:            db,
		formDecoder:   formDecoder,
	}

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

func (app App) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/login", app.login)
	router.HandlerFunc(http.MethodPost, "/login", app.loginPost)

	return router
}
