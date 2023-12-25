package main

import (
	"net/http"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/config"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	// mux.Use(WriteToConsole)
	mux.Use(CSRF)
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/check-availability", handlers.Repo.CheckAvailability)
	mux.Post("/check-availability-all", handlers.Repo.PostCheckAllAvailabilityJSON)
	mux.Post("/check-availability/{slug}", handlers.Repo.PostCheckAvailabilityJSON)
	mux.Get("/make-reservation", handlers.Repo.Reservation)
	mux.Post("/make-reservation", handlers.Repo.PostReservation)
	mux.Get("/reservation-summary", handlers.Repo.ReservationSummary)
	mux.Get("/rooms/{slug}", handlers.Repo.Room)
	mux.Get("/rooms/{slug}/book", handlers.Repo.BookRoom)
	mux.Get("/logout", handlers.Repo.Logout)

	mux.Group(func(mux chi.Router) {
		mux.Use(Guest)
		mux.Get("/login", handlers.Repo.Login)
		mux.Post("/login", handlers.Repo.PostLogin)
	})

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(Auth)
		mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		})
		mux.Get("/dashboard", handlers.Repo.AdminDashboard)
		mux.Get("/reservations/new", handlers.Repo.AdminNewReservations)
		mux.Get("/reservations/all", handlers.Repo.AdminAllReservations)
		mux.Get("/reservations/calendar", handlers.Repo.AdminReservationsCalendar)
		mux.Get("/reservations/{id}", handlers.Repo.AdminShowReservation)
		mux.Post("/reservations/{id}/edit", handlers.Repo.AdminPostEditReservation)
		mux.Post("/reservations/{id}/approve", handlers.Repo.AdminPostApproveReservation)
		mux.Post("/reservations/{id}/delete", handlers.Repo.AdminPostDeleteReservation)
	})

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
