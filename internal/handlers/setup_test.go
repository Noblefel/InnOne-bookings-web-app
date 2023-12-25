package handlers

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"text/template"
	"time"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/config"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/helpers"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/models"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/render"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
)

var app config.AppConfig
var session *scs.SessionManager
var pathToTemplates string = "./../../templates"

var functions = template.FuncMap{
	"humanDate": render.HumanDate,
}

func TestMain(m *testing.M) {
	// register the type for session
	gob.Register(models.Reservation{})
	gob.Register(models.Restriction{})
	gob.Register(models.Room{})
	gob.Register(models.RoomRestriction{})
	gob.Register(models.User{})

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

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan
	defer close(mailChan)
	listenForMail()

	templateCache, err := CreateTestTemplateCache()
	if err != nil {
		log.Fatal("Cannot get template cache", err)
	}

	app.TemplateCache = templateCache
	app.UseCache = true

	repo := NewTestRepo(&app)
	NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	os.Exit(m.Run())
}

func getRoutes() http.Handler {

	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	// mux.Use(CSRF)
	mux.Use(SessionLoad)

	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/check-availability", Repo.CheckAvailability)
	mux.Post("/check-availability-all", Repo.PostCheckAllAvailabilityJSON)
	mux.Post("/check-availability/{slug}", Repo.PostCheckAvailabilityJSON)
	mux.Get("/make-reservation", Repo.Reservation)
	mux.Post("/make-reservation", Repo.PostReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)
	mux.Get("/rooms/{slug}", Repo.Room)
	mux.Get("/rooms/{slug}/book", Repo.BookRoom)

	mux.Get("/login", Repo.Login)
	mux.Post("/login", Repo.PostLogin)
	mux.Get("/logout", Repo.Logout)

	mux.Route("/admin", func(mux chi.Router) {
		mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		})
		mux.Get("/dashboard", Repo.AdminDashboard)
		mux.Get("/reservations/new", Repo.AdminNewReservations)
		mux.Get("/reservations/all", Repo.AdminAllReservations)
		mux.Get("/reservations/calendar", Repo.AdminReservationsCalendar)
		mux.Get("/reservations/{id}", Repo.AdminShowReservation)
		mux.Post("/reservations/{id}/edit", Repo.AdminPostEditReservation)
		mux.Post("/reservations/{id}/approve", Repo.AdminPostApproveReservation)
		mux.Post("/reservations/{id}/delete", Repo.AdminPostDeleteReservation)
	})

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}

// CSRF adds csrf to every post request
func CSRF(next http.Handler) http.Handler {
	CSRF := csrf.Protect(
		[]byte("32-byte-long-auth-key"),
		csrf.SameSite(csrf.SameSiteLaxMode),
		csrf.Path("/"),
		csrf.Secure(app.InProduction),
		csrf.HttpOnly(true),
	)

	return CSRF(next)
}

// SessionLoad loads and saves session on every request
func SessionLoad(next http.Handler) http.Handler {
	return app.Session.LoadAndSave(next)
}

func CreateTestTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	// get all file named *.page.tmpl
	pages, err := filepath.Glob(pathToTemplates + "/*.page.tmpl")
	if err != nil {
		return cache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		templates, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return cache, err
		}

		layouts, err := filepath.Glob(pathToTemplates + "/*.layout.tmpl")
		if err != nil {
			return cache, err
		}

		if len(layouts) > 0 {
			templates, err = templates.ParseGlob(pathToTemplates + "/*.layout.tmpl")
			if err != nil {
				return cache, err
			}
		}

		cache[name] = templates
	}

	return cache, nil
}

func listenForMail() {
	go func() {
		for {
			_ = <-app.MailChan
		}
	}()
}
