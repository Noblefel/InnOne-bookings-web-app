package main

import (
	"net/http"
)

func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		csrfCookie, err := r.Cookie("csrf")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if csrfCookie.Value != r.FormValue("csrf") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func Auth(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.Session.GetInt(r.Context(), "auth_id") == 0 {
			app.Session.Put(r.Context(), "error", "Log in first")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Guest(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.Session.GetInt(r.Context(), "auth_id") != 0 {
			http.Redirect(w, r, r.URL.RawPath, http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
