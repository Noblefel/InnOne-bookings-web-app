package render

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/config"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/helpers"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/models"
	"github.com/gorilla/csrf"
)

var functions = template.FuncMap{
	"humanDate": HumanDate,
}

var app *config.AppConfig
var pathToTemplates string = "./templates"

func NewRenderer(ac *config.AppConfig) {
	app = ac
}

func HumanDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.CSRFToken = csrf.Token(r)
	if helpers.IsAuthenticated(r) {
		td.IsAuth = 1
	}
	return td
}

func Template(w http.ResponseWriter, r *http.Request, filename string, td *models.TemplateData) error {
	var templateCache map[string]*template.Template

	if app.UseCache {
		templateCache = app.TemplateCache
	} else {
		templateCache, _ = CreateTemplateCache()
	}

	//get requested template from cache
	template, ok := templateCache[filename]
	if !ok {
		return errors.New("Cant get template from template cache")
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td, r)

	err := template.Execute(buf, td)
	if err != nil {
		return err
	}

	// render the template
	_, err = buf.WriteTo(w)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func CreateTemplateCache() (map[string]*template.Template, error) {
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
