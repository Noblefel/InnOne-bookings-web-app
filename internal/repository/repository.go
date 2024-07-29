package repository

import (
	"time"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/types"
)

type DatabaseRepo interface {
	GetReservationById(id int) (types.Reservation, error)
	GetNewReservations() ([]types.Reservation, error)
	GetAllReservations() ([]types.Reservation, error)
	UpdateReservation(r types.Reservation) error
	DeleteReservation(id int) error
	ApproveReservation(id, processed int) error
	InsertReservation(types.Reservation) (int, error)
	InsertRoomRestriction(types.RoomRestriction) error
	SearchRoomAvailabilityByDates(start, end time.Time, roomId int) (bool, error)
	SearchAllRoomsAvailibility(start, end time.Time) ([]types.Room, error)
	GetRoomBySlug(slug string) (types.Room, error)
	GetAllRooms() ([]types.Room, error)
	GetUserById(id int) (types.User, error)
	UpdateUser(user types.User) error
	Authenticate(email, testPassword string) (int, string, error)
}
