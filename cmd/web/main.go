package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sporic/sporic/internal/models"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

type App struct {
	infoLog        *log.Logger
	errorLog       *log.Logger
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
	users          *models.UserModel
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DSN")
	addr := os.Getenv("ADDR")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	db := loadDatabase(dsn)

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour

	app := App{
		infoLog:        infoLog,
		errorLog:       errorLog,
		templateCache:  templateCache,
		users:          &models.UserModel{Db: db},
		sessionManager: sessionManager,
		formDecoder:    formDecoder,
	}

	mux := app.routes()

	srv := &http.Server{
		Addr:     addr,
		ErrorLog: errorLog,
		Handler:  mux,
	}

	infoLog.Printf("Starting server on %s", addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func loadDatabase(dsn string) *sql.DB {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db
}

func (app App) routes() http.Handler {

	dynamic := alice.New(app.sessionManager.LoadAndSave)
	router := httprouter.New()

	router.Handler(http.MethodGet, "/login", dynamic.ThenFunc(app.login))
	router.Handler(http.MethodPost, "/login", dynamic.ThenFunc(app.loginPost))
	router.Handler(http.MethodPost, "/logout", dynamic.ThenFunc(app.logout))

	return router
}
