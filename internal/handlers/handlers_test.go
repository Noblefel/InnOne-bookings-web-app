package handlers

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/models"
	"github.com/go-chi/chi/v5"
)

type postData struct {
	key   string
	value string
}

var basicTests = []struct {
	name               string
	url                string
	expectedStatusCode int
}{
	{"home-page", "/", http.StatusOK},
	{"about-page", "/about", http.StatusOK},
	{"check-availability-page", "/check-availability", http.StatusOK},
	{"invalid-page", "/this-is-invalid", http.StatusNotFound},
	{"login-page", "/login", http.StatusOK},
	{"logout", "/logout", http.StatusOK},
	{"admin-dashboard-page", "/admin/dashboard", http.StatusOK},
	{"admin-calendar-page", "/admin/reservations/calendar", http.StatusOK},
}

func getCtx(r *http.Request) context.Context {
	ctx, err := app.Session.Load(r.Context(), r.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}

	return ctx
}

type Params map[string]string

func getCtxWithParam(r *http.Request, params Params) context.Context {
	ctx := getCtx(r)
	chiCtx := chi.NewRouteContext()
	for k, v := range params {
		chiCtx.URLParams.Add(k, v)
	}
	ctx = context.WithValue(ctx, chi.RouteCtxKey, chiCtx)
	return ctx
}

// Test for simple handlers
func TestBasicHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, e := range basicTests {
		resp, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("for %v, expected %v but got %v", e.name, e.expectedStatusCode, resp.Status)
		}
	}
}

var reservation = models.Reservation{
	RoomId: 1,
	Room: models.Room{
		Id:       1,
		Slug:     "sovereigns-suite",
		RoomName: "Sovereign's Suite",
	},
	StartDate: time.Now(),
	EndDate:   time.Now().AddDate(0, 0, 1),
}

var reservationTests = []struct {
	name               string
	reservation        models.Reservation
	expectedStatusCode int
	expectedLocation   string
}{
	{
		name:               "reservation-ok",
		reservation:        reservation,
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "reservation-data-not-in-session",
		reservation:        models.Reservation{},
		expectedStatusCode: http.StatusTemporaryRedirect,
	},
}

func TestReservation(t *testing.T) {
	for _, e := range reservationTests {

		req, _ := http.NewRequest("GET", "/make-reservation", nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		if e.reservation != (models.Reservation{}) {
			app.Session.Put(ctx, "reservation", e.reservation)
		}

		handler := http.HandlerFunc(Repo.Reservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s returned response code of %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}

var postReservationTests = []struct {
	name               string
	form               url.Values
	reservation        models.Reservation
	expectedStatusCode int
}{
	{
		name: "postreservation-ok",
		form: url.Values{
			"first_name": {"John"},
			"last_name":  {"Doe"},
			"email":      {"johndoe@gmail.com"},
			"phone":      {"12345"},
		},
		reservation:        reservation,
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "postreservation-data-not-in-session",
		form: url.Values{
			"first_name": {"John"},
			"last_name":  {"Doe"},
			"email":      {"johndoe@gmail.com"},
			"phone":      {"12345"},
		},
		reservation:        models.Reservation{},
		expectedStatusCode: http.StatusTemporaryRedirect,
	},
	{
		name:               "postreservation-error-parsing-form",
		form:               nil,
		reservation:        reservation,
		expectedStatusCode: http.StatusTemporaryRedirect,
	},
	{
		name: "postreservation-failed-validation",
		form: url.Values{
			"first_name": {"x"},
			"last_name":  {"Doe"},
			"email":      {"not-email"},
		},
		reservation:        reservation,
		expectedStatusCode: http.StatusBadRequest,
	},
	{
		name: "postreservation-error-checking-room-availability",
		form: url.Values{
			"first_name": {"John"},
			"last_name":  {"Doe"},
			"email":      {"johndoe@gmail.com"},
			"phone":      {"12345"},
		},
		reservation: models.Reservation{
			RoomId: -1,
		},
		expectedStatusCode: http.StatusTemporaryRedirect,
	},
	{
		name: "postreservation-room-is-not-available",
		form: url.Values{
			"first_name": {"John"},
			"last_name":  {"Doe"},
			"email":      {"johndoe@gmail.com"},
			"phone":      {"12345"},
		},
		reservation: models.Reservation{
			RoomId:    1,
			StartDate: time.Now().AddDate(-100, 0, 0),
			EndDate:   time.Now().AddDate(100, 0, 0),
		},
		expectedStatusCode: http.StatusConflict,
	},
	{
		name: "postreservation-error-inserting-reservation",
		form: url.Values{
			"first_name": {"John"},
			"last_name":  {"Doe"},
			"email":      {"johndoe@gmail.com"},
			"phone":      {"12345"},
		},
		reservation: models.Reservation{
			RoomId: 1,
			Room: models.Room{
				Slug: "non-existent-room",
			},
		},
		expectedStatusCode: http.StatusTemporaryRedirect,
	},
	{
		name: "postreservation-error-inserting-restriction",
		form: url.Values{
			"first_name": {"John"},
			"last_name":  {"Doe"},
			"email":      {"johndoe@gmail.com"},
			"phone":      {"12345"},
		},
		reservation: models.Reservation{
			RoomId: 1,
			Room: models.Room{
				Slug: "get-invalid-id-for-restriction",
			},
		},
		expectedStatusCode: http.StatusTemporaryRedirect,
	},
}

func TestPostReservation(t *testing.T) {
	for _, e := range postReservationTests {

		var req *http.Request
		if e.form != nil {
			req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(e.form.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/make-reservation", nil)
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		if e.reservation != (models.Reservation{}) {
			app.Session.Put(ctx, "reservation", e.reservation)
		}

		handler := http.HandlerFunc(Repo.PostReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s returned response code of %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}

var reservationSummaryTests = []struct {
	name               string
	reservation        models.Reservation
	expectedStatusCode int
}{
	{
		name:               "reservationsummary-ok",
		reservation:        reservation,
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "reservationsummary-data-not-in-session",
		reservation:        models.Reservation{},
		expectedStatusCode: http.StatusTemporaryRedirect,
	},
}

func TestReservationSummary(t *testing.T) {
	for _, e := range reservationSummaryTests {
		req, _ := http.NewRequest("GET", "/reservation-summary", nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		if e.reservation != (models.Reservation{}) {
			app.Session.Put(ctx, "reservation", e.reservation)
		}

		handler := http.HandlerFunc(Repo.ReservationSummary)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s returned response code of %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}

var checkAllAvailabilityJSONTests = []struct {
	name               string
	form               url.Values
	expectedStatusCode int
}{
	{
		name: "checkAllAvailabilityJSON-ok",
		form: url.Values{
			"start_date": {"2023-01-02"},
			"end_date":   {"2023-01-03"},
		},
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "checkAllAvailabilityJSON-error-parsing-form",
		form:               nil,
		expectedStatusCode: http.StatusBadRequest,
	},
	{
		name: "checkAllAvailabilityJSON-error-parsing-start-date",
		form: url.Values{
			"start_date": {"invalid"},
			"end_date":   {"2023-01-03"},
		},
		expectedStatusCode: http.StatusBadRequest,
	},
	{
		name: "checkAllAvailabilityJSON-error-parsing-end-date",
		form: url.Values{
			"start_date": {"2023-01-02"},
			"end_date":   {"invalid"},
		},
		expectedStatusCode: http.StatusBadRequest,
	},
	{
		name: "checkAllAvailabilityJSON-error-search-all-room-availability",
		form: url.Values{
			"start_date": {"1500-01-01"},
			"end_date":   {"2023-01-03"},
		},
		expectedStatusCode: http.StatusInternalServerError,
	},
	{
		name: "checkAllAvailabilityJSON-error-marshalling-json",
		form: url.Values{
			"start_date": {"2023-01-02"},
			"end_date":   {"2023-01-03"},
			"TEST_ERROR": {"x"},
		},
		expectedStatusCode: http.StatusInternalServerError,
	},
}

func TestPostCheckAllAvailabilityJSON(t *testing.T) {
	for _, e := range checkAllAvailabilityJSONTests {

		var req *http.Request
		if e.form != nil {
			req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(e.form.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/make-reservation", nil)
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.PostCheckAllAvailabilityJSON)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s returned response code of %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}

var room = models.Room{
	Id:       1,
	RoomName: "Sovereign's Suite",
	Slug:     "sovereigns-suite",
}

var checkAvailabilityJSONTests = []struct {
	name               string
	form               url.Values
	room               models.Room
	urlSlug            string
	expectedStatusCode int
}{
	{
		name: "checkAllAvailabilityJSON-ok",
		form: url.Values{
			"start_date": {"2023-01-02"},
			"end_date":   {"2023-01-03"},
		},
		room:               room,
		urlSlug:            room.Slug,
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "checkAllAvailabilityJSON-error-parsing-form",
		form:               nil,
		expectedStatusCode: http.StatusBadRequest,
	},
	{
		name: "checkAllAvailabilityJSON-error-parsing-start-date",
		form: url.Values{
			"start_date": {"invalid"},
			"end_date":   {"2023-01-03"},
		},
		expectedStatusCode: http.StatusBadRequest,
	},
	{
		name: "checkAllAvailabilityJSON-error-parsing-end-date",
		form: url.Values{
			"start_date": {"2023-01-02"},
			"end_date":   {"invalid"},
		},
		expectedStatusCode: http.StatusBadRequest,
	},
	{
		name: "checkAllAvailabilityJSON-data-not-in-session",
		form: url.Values{
			"start_date": {"2023-01-02"},
			"end_date":   {"2023-01-03"},
		},
		room:               models.Room{},
		expectedStatusCode: http.StatusInternalServerError,
	},
	{
		name: "checkAllAvailabilityJSON-url-slug-does-not-match",
		form: url.Values{
			"start_date": {"2023-01-02"},
			"end_date":   {"2023-01-03"},
		},
		room:               room,
		urlSlug:            "invalid",
		expectedStatusCode: http.StatusInternalServerError,
	},
	{
		name: "checkAllAvailabilityJSON-error-checking-room-availability",
		form: url.Values{
			"start_date": {"2023-01-02"},
			"end_date":   {"2023-01-03"},
		},
		room:               models.Room{Id: -1, Slug: room.Slug},
		urlSlug:            room.Slug,
		expectedStatusCode: http.StatusInternalServerError,
	},
	{
		name: "checkAllAvailabilityJSON-room-is-not-available",
		form: url.Values{
			"start_date": {"1500-01-02"},
			"end_date":   {"2500-01-03"},
		},
		room:               room,
		urlSlug:            room.Slug,
		expectedStatusCode: http.StatusNotFound,
	},
	{
		name: "checkAllAvailabilityJSON-error-marshalling-json",
		form: url.Values{
			"start_date": {"2023-01-02"},
			"end_date":   {"2023-01-03"},
			"TEST_ERROR": {"x"},
		},
		room:               room,
		urlSlug:            room.Slug,
		expectedStatusCode: http.StatusInternalServerError,
	},
}

func TestPostCheckAvailabilityJSON(t *testing.T) {
	for _, e := range checkAvailabilityJSONTests {

		var req *http.Request
		if e.form != nil {
			req, _ = http.NewRequest("POST", "/check-availability/{slug}", strings.NewReader(e.form.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/check-availability/{slug}", nil)
		}

		ctx := getCtxWithParam(req, Params{"slug": e.urlSlug})

		if e.room != (models.Room{}) {
			app.Session.Put(ctx, "room", e.room)
		}

		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.PostCheckAvailabilityJSON)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s returned response code of %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}

var roomTests = []struct {
	name               string
	urlSlug            string
	expectedStatusCode int
}{
	{
		name:               "room-ok",
		urlSlug:            room.Slug,
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "room-error-getting-room-by-slug",
		urlSlug:            "error",
		expectedStatusCode: http.StatusTemporaryRedirect,
	},
	{
		name:               "room-room-not-found",
		urlSlug:            "non-existent",
		expectedStatusCode: http.StatusSeeOther,
	},
}

func TestRoom(t *testing.T) {
	for _, e := range roomTests {
		req, _ := http.NewRequest("GET", "/rooms/{slug}", nil)
		ctx := getCtxWithParam(req, Params{"slug": e.urlSlug})
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.Room)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s returned response code of %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}

var bookRoomTests = []struct {
	name               string
	urlSlug            string
	urlParams          string
	expectedStatusCode int
}{
	{
		name:               "bookRoom-ok",
		urlSlug:            room.Slug,
		urlParams:          "start_date=2020-01-02&end_date=2020-01-03",
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name:               "bookRoom-error-parsing-start-date",
		urlSlug:            room.Slug,
		urlParams:          "start_date=invalid&end_date=2020-01-03",
		expectedStatusCode: http.StatusTemporaryRedirect,
	},
	{
		name:               "bookRoom-error-parsing-end-date",
		urlSlug:            room.Slug,
		urlParams:          "start_date=2020-01-02&end_date=invalid",
		expectedStatusCode: http.StatusTemporaryRedirect,
	},
	{
		name:               "bookRoom-error-getting-room-by-slug",
		urlSlug:            "error",
		urlParams:          "start_date=2020-01-02&end_date=2020-01-03",
		expectedStatusCode: http.StatusTemporaryRedirect,
	},
	{
		name:               "bookRoom-room-not-found",
		urlSlug:            "non-existent",
		urlParams:          "start_date=2020-01-02&end_date=2020-01-03",
		expectedStatusCode: http.StatusSeeOther,
	},
}

func TestBookRoom(t *testing.T) {
	for _, e := range bookRoomTests {
		req, _ := http.NewRequest("GET", "/rooms/{slug}/book?"+e.urlParams, nil)
		ctx := getCtxWithParam(req, Params{"slug": e.urlSlug})
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.BookRoom)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s returned response code of %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}

var postLoginTests = []struct {
	name               string
	form               url.Values
	expectedStatusCode int
}{
	{
		name: "postLogin-ok",
		form: url.Values{
			"email":    {"test@example.com"},
			"password": {"password"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name:               "postLogin-error-parsing-form",
		form:               nil,
		expectedStatusCode: http.StatusTemporaryRedirect,
	},
	{
		name: "postLogin-failed-validation",
		form: url.Values{
			"email": {"x"},
		},
		expectedStatusCode: http.StatusBadRequest,
	},
	{
		name: "postLogin-error-invalid-credentials",
		form: url.Values{
			"email":    {"test@example.com"},
			"password": {"invalid-credentials"},
		},
		expectedStatusCode: http.StatusUnauthorized,
	},
	{
		name: "postLogin-error-authenticating",
		form: url.Values{
			"email":    {"test@example.com"},
			"password": {"error"},
		},
		expectedStatusCode: http.StatusTemporaryRedirect,
	},
}

func TestPostLogin(t *testing.T) {
	for _, e := range postLoginTests {

		var req *http.Request
		if e.form != nil {
			req, _ = http.NewRequest("POST", "/login", strings.NewReader(e.form.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/login", nil)
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.PostLogin)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s returned response code of %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}

var AdminGetReservationsTests = []struct {
	name               string
	url                string
	expectedStatusCode int
}{
	{
		name:               "adminNewReservations-ok",
		url:                "/admin/reservations/new",
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "adminNewReservations-error-getting-new-reservations",
		url:                "/admin/reservations/new-TEST_ERROR",
		expectedStatusCode: http.StatusInternalServerError,
	},
	{
		name:               "adminAllReservations-ok",
		url:                "/admin/reservations/all",
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "adminAllReservations-error-getting-all-reservations",
		url:                "/admin/reservations/all-TEST_ERROR",
		expectedStatusCode: http.StatusInternalServerError,
	},
}

func TestAdminNewReservations(t *testing.T) {
	for _, e := range AdminGetReservationsTests {

		req, _ := http.NewRequest("GET", e.url, nil)

		ctx := getCtx(req)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		var handler http.HandlerFunc
		if strings.Contains(e.url, "new") {
			handler = http.HandlerFunc(Repo.AdminNewReservations)
		} else {
			handler = http.HandlerFunc(Repo.AdminAllReservations)
		}

		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s returned response code of %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}

var adminShowReservationTests = []struct {
	name               string
	reservationId      string
	expectedStatusCode int
}{
	{
		name:               "adminShowReservations-ok",
		reservationId:      "1",
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "adminShowReservations-error-url-param-id",
		reservationId:      "x",
		expectedStatusCode: http.StatusInternalServerError,
	},
	{
		name:               "adminShowReservations-error-getting-reservation",
		reservationId:      "-1",
		expectedStatusCode: http.StatusInternalServerError,
	},
	{
		name:               "adminShowReservations-error-reservation-not-found",
		reservationId:      "0",
		expectedStatusCode: http.StatusTemporaryRedirect,
	},
}

func TestAdminShowReservations(t *testing.T) {
	for _, e := range adminShowReservationTests {

		req, _ := http.NewRequest("GET", "/admin/reservations/{id}", nil)
		ctx := getCtxWithParam(req, Params{"id": e.reservationId})
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.AdminShowReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s returned response code of %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}

var adminPostEditReservationTests = []struct {
	name               string
	reservationId      string
	form               url.Values
	expectedStatusCode int
}{
	{
		name:          "adminPostEditReservation-ok",
		reservationId: "1",
		form: url.Values{
			"first_name": {"John"},
			"last_name":  {"Doe"},
			"email":      {"test@example.com"},
			"phone":      {"12412"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name:               "adminPostEditReservation-error-url-param-id",
		reservationId:      "x",
		form:               nil,
		expectedStatusCode: http.StatusInternalServerError,
	},
	{
		name:               "adminPostEditReservation-error-parsing-form",
		reservationId:      "1",
		form:               nil,
		expectedStatusCode: http.StatusInternalServerError,
	},
	{
		name:               "adminPostEditReservation-error-getting-reservation",
		reservationId:      "-1",
		form:               url.Values{},
		expectedStatusCode: http.StatusInternalServerError,
	},
	{
		name:          "adminPostEditReservation-failed-validation",
		reservationId: "1",
		form: url.Values{
			"email": {"x"},
		},
		expectedStatusCode: http.StatusBadRequest,
	},
	{
		name:          "adminPostEditReservation-error-updating-reservation",
		reservationId: "1",
		form: url.Values{
			"first_name": {"error"},
			"last_name":  {"Doe"},
			"email":      {"test@example.com"},
			"phone":      {"12412"},
		},
		expectedStatusCode: http.StatusInternalServerError,
	},
}

func TestAdminPostEditReservation(t *testing.T) {
	for _, e := range adminPostEditReservationTests {

		url := "/admin/reservations/{id}/edit"

		var req *http.Request
		if e.form != nil {
			req, _ = http.NewRequest("POST", url, strings.NewReader(e.form.Encode()))
		} else {
			req, _ = http.NewRequest("POST", url, nil)
		}

		ctx := getCtxWithParam(req, Params{"id": e.reservationId})
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AdminPostEditReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s returned response code of %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}

var adminPostApproveReservationTests = []struct {
	name               string
	reservationId      string
	form               url.Values
	expectedStatusCode int
}{
	{
		name:          "adminPostApproveReservation-ok",
		reservationId: "1",
		form: url.Values{
			"is-approved": {"1"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name:               "adminPostApproveReservation-error-url-param-id",
		reservationId:      "x",
		form:               nil,
		expectedStatusCode: http.StatusInternalServerError,
	},
	{
		name:               "adminPostApproveReservation-error-parsing-form",
		reservationId:      "1",
		form:               nil,
		expectedStatusCode: http.StatusInternalServerError,
	},
	{
		name:          "adminPostApproveReservation-error-approving-reservation-0",
		reservationId: "0",
		form: url.Values{
			"is-approved": {"0"},
		},
		expectedStatusCode: http.StatusInternalServerError,
	},
	{
		name:          "adminPostApproveReservation-error-approving-reservation-1",
		reservationId: "0",
		form: url.Values{
			"is-approved": {"1"},
		},
		expectedStatusCode: http.StatusInternalServerError,
	},
}

func TestAdminPostApproveReservation(t *testing.T) {
	for _, e := range adminPostApproveReservationTests {

		url := "/admin/reservations/{id}/approve"

		var req *http.Request
		if e.form != nil {
			req, _ = http.NewRequest("POST", url, strings.NewReader(e.form.Encode()))
		} else {
			req, _ = http.NewRequest("POST", url, nil)
		}

		ctx := getCtxWithParam(req, Params{"id": e.reservationId})
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AdminPostApproveReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s returned response code of %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}

var AdminPostDeleteReservationTests = []struct {
	name               string
	reservationId      string
	expectedStatusCode int
}{
	{
		name:               "adminPostDeleteReservation-ok",
		reservationId:      "1",
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name:               "adminPostDeleteReservation-error-url-param-id",
		reservationId:      "x",
		expectedStatusCode: http.StatusInternalServerError,
	},
	{
		name:               "adminPostDeleteReservation-error-deleting-reservation",
		reservationId:      "0",
		expectedStatusCode: http.StatusInternalServerError,
	},
}

func TestAdminPostDeleteReservation(t *testing.T) {
	for _, e := range AdminPostDeleteReservationTests {

		req, _ := http.NewRequest("POST", "/admin/reservations/{id}/approve", nil)

		ctx := getCtxWithParam(req, Params{"id": e.reservationId})
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AdminPostDeleteReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s returned response code of %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}
