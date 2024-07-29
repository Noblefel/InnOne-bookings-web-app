package main

import (
	"database/sql"
	"embed"
	"encoding/gob"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/handlers"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/render"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/repository"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/types"
	"github.com/alexedwards/scs/v2"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed "templates/*"
var templateFS embed.FS

//go:embed "static/*"
var staticFS embed.FS

const port = 8080

var app *types.App

func main() {
	dbHost := flag.String("dbhost", "localhost", "database host")
	dbPort := flag.String("dbport", "5432", "database port")
	dbName := flag.String("dbname", "bookings", "database name")
	dbUser := flag.String("dbuser", "postgres", "database user")
	dbPW := flag.String("dbpw", "", "database password")
	flag.Parse()

	// register the type for session
	gob.Register(types.Reservation{})
	gob.Register(types.Restriction{})
	gob.Register(types.Room{})
	gob.Register(types.RoomRestriction{})
	gob.Register(types.User{})

	app = &types.App{
		MailChan: make(chan types.MailData),
		// InfoLog:  log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		// ErrorLog: log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		Session: scs.New(),
	}
	defer close(app.MailChan)
	listenForMail()

	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s",
		*dbHost,
		*dbPort,
		*dbName,
		*dbUser,
		*dbPW,
	)

	db, err := connectDb(dsn)
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewPostgresRepo(db, app)
	// renderer := render.NewNoCache("templates")
	templateFS, _ := fs.Sub(templateFS, "templates")
	renderer, err := render.New(templateFS)
	if err != nil {
		log.Fatal(err)
	}
	handlers := handlers.New(app, repo, renderer)

	log.Println("Starting server at port:", port)

	serve := &http.Server{
		Addr:    fmt.Sprint("localhost:", port),
		Handler: routes(handlers),
	}

	if err = serve.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func connectDb(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
