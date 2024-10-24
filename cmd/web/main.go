package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sporic/sporic/internal/mailer"
	"github.com/sporic/sporic/internal/models"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/robfig/cron/v3"
)

type App struct {
	infoLog        *log.Logger
	errorLog       *log.Logger
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
	users          *models.UserModel
	applications   *models.ApplicationModel
	notifications  *models.NotificationModel
	payments       *models.PaymentModel
	mailer         mailer.Mailer
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

	smtp_host := os.Getenv("SMTP_HOST")
	smtp_port, err := strconv.ParseUint(os.Getenv("SMTP_PORT"), 10, 16)
	if err != nil {
		log.Fatalf("Error parsing SMTP_PORT: %v", err)
	}
	smpt_username := os.Getenv("SMTP_USER")
	smpt_password := os.Getenv("SMTP_PASS")
	smpt_sender := os.Getenv("SMTP_SENDER")

	mailer := mailer.New(smtp_host, int(smtp_port), smpt_username, smpt_password, smpt_sender)

	app := App{
		infoLog:        infoLog,
		errorLog:       errorLog,
		templateCache:  templateCache,
		users:          &models.UserModel{Db: db},
		applications:   &models.ApplicationModel{Db: db},
		payments:       &models.PaymentModel{Db: db},
		notifications:  &models.NotificationModel{Db: db},
		sessionManager: sessionManager,
		formDecoder:    formDecoder,
		mailer:         mailer,
	}

	c := cron.New()

	_, err = c.AddFunc("@daily", app.checkForDelayes)
	if err != nil {
		log.Fatalf("Error adding cron job: %v", err)
	}

	c.Start()

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
	if err := db.Ping(); err != nil {
		log.Fatalln(err)
	}

	return db
}

func (app App) routes() http.Handler {

	dynamic := alice.New(app.sessionManager.LoadAndSave, app.authenticateMiddleware)
	router := httprouter.New()

	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/login", dynamic.ThenFunc(app.login))
	router.Handler(http.MethodPost, "/login", dynamic.ThenFunc(app.loginPost))
	router.Handler(http.MethodPost, "/logout", dynamic.ThenFunc(app.logout))
	router.Handler(http.MethodGet, "/home", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/admin_home", dynamic.ThenFunc(app.admin_home))
	router.Handler(http.MethodGet, "/faculty_home", dynamic.ThenFunc(app.faculty_home))
	router.Handler(http.MethodGet, "/new_application", dynamic.ThenFunc(app.new_application))
	router.Handler(http.MethodPost, "/new_application", dynamic.ThenFunc(app.new_application_post))
	router.Handler(http.MethodGet, "/faculty/view_application/:refno", dynamic.ThenFunc(app.faculty_view_application))
	router.Handler(http.MethodPost, "/faculty/view_application/:refno", dynamic.ThenFunc(app.faculty_view_application))
	router.Handler(http.MethodGet, "/admin/view_application/:refno", dynamic.ThenFunc(app.admin_view_application))
	router.Handler(http.MethodPost, "/admin/view_application/:refno", dynamic.ThenFunc(app.admin_view_application))
	router.Handler(http.MethodGet, "/download/:folder/:doc_id/:doc_type", dynamic.ThenFunc(app.download))
	router.Handler(http.MethodGet, "/accounts_home", dynamic.ThenFunc(app.accounts_home))
	router.Handler(http.MethodPost, "/accounts_home", dynamic.ThenFunc(app.accounts_home))
	router.Handler(http.MethodGet, "/notifications", dynamic.ThenFunc(app.GetNotifications))
	router.Handler(http.MethodGet, "/profile", dynamic.ThenFunc(app.profile))
	router.Handler(http.MethodGet, "/excel", dynamic.ThenFunc(app.excel))
	router.Handler(http.MethodGet, "/provc_home", dynamic.ThenFunc(app.provc_home))
	router.Handler(http.MethodGet, "/provc/view_application/:refno", dynamic.ThenFunc(app.provc_view_application))
	router.Handler(http.MethodPost, "/provc/view_application/:refno", dynamic.ThenFunc(app.provc_view_application))
	return router
}
