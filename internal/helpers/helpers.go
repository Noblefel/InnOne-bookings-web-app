package helpers

import (
	"encoding/json"
	"net/http"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/config"
)

var app *config.AppConfig

type JSONResponse struct {
	Ok      bool        `json:"ok"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// NewHelpers sets app config for helpers
func NewHelpers(a *config.AppConfig) {
	app = a
}

func ClientError(w http.ResponseWriter, status int) {
	app.InfoLog.Println("Client error with status of", status)
	http.Error(w, http.StatusText(status), status)
}

func ServerError(w http.ResponseWriter, err error) {
	// trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	// app.ErrorLog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func ServerErrorJSON(w http.ResponseWriter, code int, err error, msg string) {
	// if err != nil {
	// 	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	// 	app.ErrorLog.Println(trace)
	// }

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	message := http.StatusText(code)
	if msg != "" {
		message = msg
	}

	json.NewEncoder(w).Encode(JSONResponse{
		Ok:      false,
		Message: message,
	})
}

func IsAuthenticated(r *http.Request) bool {
	return app.Session.Exists(r.Context(), "user_id")
}
