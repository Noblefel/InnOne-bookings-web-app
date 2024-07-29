package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/types"
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

type mockRepo struct{}

func NewMockRepo() *mockRepo { return &mockRepo{} }

func (m *mockRepo) GetReservationById(id int) (types.Reservation, error) {
	var reservation types.Reservation

	if id < 0 {
		return reservation, errors.New("some error")
	}

	if id == 0 {
		return reservation, sql.ErrNoRows
	}

	return reservation, nil
}

func (m *mockRepo) GetNewReservations() ([]types.Reservation, error) {
	var reservations []types.Reservation
	return reservations, nil
}

func (m *mockRepo) GetAllReservations() ([]types.Reservation, error) {
	var reservations []types.Reservation
	return reservations, nil
}

func (m *mockRepo) UpdateReservation(r types.Reservation) error {
	if r.FirstName == "error" {
		return errors.New("Some errors")
	}

	return nil
}

func (m *mockRepo) DeleteReservation(id int) error {
	if id == 0 {
		return errors.New("Some errors")
	}

	return nil
}

func (m *mockRepo) ApproveReservation(id, processed int) error {
	if id == 0 {
		return errors.New("Some errors")
	}

	return nil
}

func (m *mockRepo) InsertReservation(r types.Reservation) (int, error) {
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

func (m *mockRepo) InsertRoomRestriction(rr types.RoomRestriction) error {
	if rr.RoomId <= 0 || rr.RoomId > len(roomSlugs) || rr.ReservationId <= 0 {
		return errors.New("Some error")
	}

	return nil
}

func (m *mockRepo) SearchRoomAvailabilityByDates(start, end time.Time, roomId int) (bool, error) {
	if roomId <= 0 || roomId > len(roomSlugs) {
		return false, errors.New("Some error")
	}

	if start.After(end) {
		return false, nil
	}

	return true, nil
}

func (m *mockRepo) SearchAllRoomsAvailibility(start, end time.Time) ([]types.Room, error) {
	if start.Before(time.Now().AddDate(-100, 0, 0)) {
		return []types.Room{}, errors.New("Some error")
	}

	return []types.Room{}, nil
}

func (m *mockRepo) GetRoomBySlug(slug string) (types.Room, error) {
	if slug == "error" {
		return types.Room{}, errors.New("some error")
	}

	for _, s := range roomSlugs {
		if s == slug {
			return types.Room{RoomName: "Not Empty"}, nil
		}
	}

	return types.Room{}, nil
}

func (m *mockRepo) GetAllRooms() ([]types.Room, error) {
	return []types.Room{}, nil
}

func (m *mockRepo) GetUserById(id int) (types.User, error) {
	return types.User{}, nil
}

func (m *mockRepo) UpdateUser(user types.User) error {
	return nil
}

func (m *mockRepo) Authenticate(email, testPassword string) (int, string, error) {
	return 1, "", nil
}
