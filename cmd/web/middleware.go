package main

import (
	"net/http"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/helpers"
	"github.com/gorilla/csrf"
)

// func WriteToConsole(next http.Handler) http.Handler {

// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Println("Page visited")
// 		next.ServeHTTP(w, r)
// 	})
// }

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

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !helpers.IsAuthenticated(r) {
			app.Session.Put(r.Context(), "error", "Log in first")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Guest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if helpers.IsAuthenticated(r) {
			http.Redirect(w, r, r.URL.RawPath, http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
