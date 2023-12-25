package render

import (
	"net/http"
	"testing"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/models"
)

func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	session.Put(r.Context(), "flash", "test")
	result := AddDefaultData(&td, r)

	if len(result.Flash) == 0 {
		t.Error("Flash value not found in session")
	}
}

func TestRenderTemplate(t *testing.T) {
	pathToTemplates = "./../../templates"

	tc, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}

	app.TemplateCache = tc

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	err = Template(&myWriter{}, r, "home.page.tmpl", &models.TemplateData{})
	if err != nil {
		t.Error(err)
	}

	err = Template(&myWriter{}, r, "didnt-exist.page.tmpl", &models.TemplateData{})
	if err == nil {
		t.Error("Rendered template that didnt exists")
	}

}

func TestCreateTemplateCache(t *testing.T) {
	pathToTemplates = "./../../templates"

	_, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}
}

func getSession() (*http.Request, error) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		return nil, err
	}

	ctx, _ := session.Load(r.Context(), r.Header.Get("X-Session"))

	r = r.WithContext(ctx)

	return r, nil
}

func TestNewTemplate(t *testing.T) {
	NewRenderer(app)
}
