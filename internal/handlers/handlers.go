package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/types"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/forms"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/render"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type Handlers struct {
	*types.App
	repo     repository.DatabaseRepo
	renderer render.Renderer
}

type JSONResponse struct {
	Ok      bool        `json:"ok"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func ServerErrorJSON(w http.ResponseWriter, code int, err error, msg string) {
	log.Println(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if msg == "" {
		msg = http.StatusText(code)
	}

	json.NewEncoder(w).Encode(JSONResponse{
		Ok:      false,
		Message: msg,
	})
}

// view Wraps renderer.View with global state
func (app *Handlers) view(w http.ResponseWriter, r *http.Request, tmpl string, page map[string]any) {
	if page == nil {
		page = make(map[string]any)
	}

	var d types.TemplateData
	d.Flash = app.Session.PopString(r.Context(), "flash")
	d.Error = app.Session.PopString(r.Context(), "error")
	d.AuthId = app.Session.GetInt(r.Context(), "auth_id")
	d.Page = page

	if err := app.renderer.View(w, tmpl, &d); err != nil {
		log.Println(err)
	}
}

func csrfToken(w http.ResponseWriter) string {
	csrf := make([]byte, 15)
	rand.Read(csrf)
	token := base64.URLEncoding.EncodeToString(csrf)

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf",
		Value:    token,
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
	})

	return token
}

func New(c *types.App, repo repository.DatabaseRepo, renderer render.Renderer) *Handlers {
	return &Handlers{c, repo, renderer}
}

func (app *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	app.view(w, r, "home.page.tmpl", nil)
}

func (app *Handlers) About(w http.ResponseWriter, r *http.Request) {
	app.view(w, r, "about.page.tmpl", nil)
}

func (app *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	app.view(w, r, "login.page.tmpl", map[string]any{
		"form": forms.New(nil),
		"csrf": csrfToken(w),
	})
}

func (app *Handlers) PostLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.Session.Put(r.Context(), "error", "Can't pass form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	newToken := csrfToken(w)
	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		app.Session.Put(r.Context(), "error", "Some fields are invalid")
		w.WriteHeader(http.StatusBadRequest)
		app.view(w, r, "login.page.tmpl", map[string]any{
			"errors": form.Errors,
			"csrf":   newToken,
		})
		return
	}

	id, _, err := app.repo.Authenticate(r.Form.Get("email"), r.Form.Get("password"))
	if err != nil {
		if errors.Is(bcrypt.ErrMismatchedHashAndPassword, err) {
			app.Session.Put(r.Context(), "error", "Invalid credentials")
			w.WriteHeader(http.StatusUnauthorized)
			app.view(w, r, "login.page.tmpl", map[string]any{"errors": form.Errors})
			return
		}

		app.Session.Put(r.Context(), "error", "Something went wrong, cannot authenticate")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	app.Session.Put(r.Context(), "auth_id", id)
	app.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	_ = app.Session.Destroy(r.Context())
	_ = app.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
