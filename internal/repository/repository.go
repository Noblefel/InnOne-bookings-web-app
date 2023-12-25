package repository

import (
	"time"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/models"
)

type DatabaseRepo interface {
	AllUsers() bool
	GetReservationById(id int) (models.Reservation, error)
	GetNewReservations() ([]models.Reservation, error)
	GetAllReservations() ([]models.Reservation, error)
	UpdateReservation(r models.Reservation) error
	DeleteReservation(id int) error
	ApproveReservation(id, processed int) error
	InsertReservation(models.Reservation) (int, error)
	InsertRoomRestriction(models.RoomRestriction) error
	SearchRoomAvailabilityByDates(start, end time.Time, roomId int) (bool, error)
	SearchAllRoomsAvailibility(start, end time.Time) ([]models.Room, error)
	GetRoomBySlug(slug string) (models.Room, error)
	GetAllRooms() ([]models.Room, error)
	GetUserById(id int) (models.User, error)
	UpdateUser(user models.User) error
	Authenticate(email, testPassword string) (int, string, error)
}
