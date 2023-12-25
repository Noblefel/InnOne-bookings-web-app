package render

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/config"
	"github.com/Noblefel/InnOne-bookings-web-app/internal/models"
	"github.com/alexedwards/scs/v2"
)

var session *scs.SessionManager
var testApp config.AppConfig

func TestMain(m *testing.M) {
	// register the type for session
	gob.Register(models.Reservation{})

	//  change to true when in production
	testApp.InProduction = false

	testApp.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	testApp.ErrorLog = log.New(os.Stdout, "Error\t", log.Ldate|log.Ltime|log.Lshortfile)

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = false

	testApp.Session = session

	app = &testApp

	os.Exit(m.Run())
}

type myWriter struct{}

func (w *myWriter) Header() http.Header {
	return http.Header{}
}

func (w *myWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func (w *myWriter) WriteHeader(i int) {}
