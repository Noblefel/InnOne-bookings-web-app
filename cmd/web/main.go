package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/config"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/driver"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/handlers"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/helpers"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/models"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/render"
	"github.com/alexedwards/scs/v2"
	"github.com/joho/godotenv"
)

const port = 8080

var app config.AppConfig

func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()

	defer close(app.MailChan)
	log.Println("Starting mail listener...")
	listenForMail()

	fmt.Println("Starting server at port:", port)

	serve := &http.Server{
		Addr:    fmt.Sprint("localhost:", port),
		Handler: routes(&app),
	}

	err = serve.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}

func run() (*driver.DB, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	// register the type for session
	gob.Register(models.Reservation{})
	gob.Register(models.Restriction{})
	gob.Register(models.Room{})
	gob.Register(models.RoomRestriction{})
	gob.Register(models.User{})

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	//  change to true when in production
	app.InProduction = false

	app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "Error\t", log.Ldate|log.Ltime|log.Lshortfile)

	session := scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	// connect to database
	log.Println("Connecting to the database...")
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
	)

	db, err := driver.ConnectSQL(dsn)
	if err != nil {
		log.Fatal("Cannot connect to database")
		return nil, err
	}
	log.Println("Connected to the database")

	templateCache, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Cannot get template cache")
		return nil, err
	}

	app.TemplateCache = templateCache

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, nil
}
