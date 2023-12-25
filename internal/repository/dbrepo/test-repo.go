package dbrepo

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var roomSlugs = []string{
	"sovereigns-suite",
	"queens-lounge",
	"lords-lair",
	"duchess-domain",
	"prince-manor",
	"princess-stay",
	"knights-quarter",
	"heralds-retreat",
}

func (m *testDBRepo) AllUsers() bool {
	return true
}

func (m *testDBRepo) GetReservationById(id int) (models.Reservation, error) {
	var reservation models.Reservation

	if id < 0 {
		return reservation, errors.New("Some error")
	}

	if id == 0 {
		return reservation, sql.ErrNoRows
	}

	return reservation, nil
}

func (m *testDBRepo) GetNewReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation
	return reservations, nil
}

func (m *testDBRepo) GetAllReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation
	return reservations, nil
}

func (m *testDBRepo) UpdateReservation(r models.Reservation) error {
	if r.FirstName == "error" {
		return errors.New("Some errors")
	}

	return nil
}

func (m *testDBRepo) DeleteReservation(id int) error {
	if id == 0 {
		return errors.New("Some errors")
	}

	return nil
}

func (m *testDBRepo) ApproveReservation(id, processed int) error {
	if id == 0 {
		return errors.New("Some errors")
	}

	return nil
}

func (m *testDBRepo) InsertReservation(r models.Reservation) (int, error) {
	if r.Room.Slug == "get-invalid-id-for-restriction" {
		return -1000, nil // to give InsertRoomRestriction() an error
	}

	for _, s := range roomSlugs {
		if s == r.Room.Slug {
			return 1, nil
		}
	}

	return 0, errors.New("Some error")
}

func (m *testDBRepo) InsertRoomRestriction(rr models.RoomRestriction) error {
	if rr.RoomId <= 0 || rr.RoomId > len(roomSlugs) || rr.ReservationId <= 0 {
		return errors.New("Some error")
	}

	return nil
}

func (m *testDBRepo) SearchRoomAvailabilityByDates(start, end time.Time, roomId int) (bool, error) {
	layout := "2006-01-02"
	start2, _ := time.Parse(layout, "2020-01-01")
	end2, _ := time.Parse(layout, "2020-12-30")

	if roomId <= 0 || roomId > len(roomSlugs) {
		return false, errors.New("Some error")
	}

	if start.Before(end2) && end.After(start2) {
		return false, nil // passed but not available
	}

	return true, nil // passed and available
}

func (m *testDBRepo) SearchAllRoomsAvailibility(start, end time.Time) ([]models.Room, error) {
	if start.Before(time.Now().AddDate(-100, 0, 0)) {
		return []models.Room{}, errors.New("Some error")
	}

	return []models.Room{}, nil
}

func (m *testDBRepo) GetRoomBySlug(slug string) (models.Room, error) {

	if slug == "error" {
		return models.Room{}, errors.New("some error")
	}

	for _, s := range roomSlugs {
		if s == slug {
			return models.Room{RoomName: "Not Empty"}, nil
		}
	}

	return models.Room{}, nil
}

func (m *testDBRepo) GetAllRooms() ([]models.Room, error) {
	return []models.Room{}, nil
}

func (m *testDBRepo) GetUserById(id int) (models.User, error) {
	return models.User{}, nil
}

func (m *testDBRepo) UpdateUser(user models.User) error {
	return nil
}

func (m *testDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	if testPassword == "invalid-credentials" {
		return 0, "", bcrypt.ErrMismatchedHashAndPassword
	}

	if testPassword == "error" {
		return 0, "", errors.New("Some unexpected error")
	}

	return 1, "", nil
}
