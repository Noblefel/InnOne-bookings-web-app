package types

import (
	"time"

	"github.com/alexedwards/scs/v2"
)

type TemplateData struct {
	Page   map[string]any
	Flash  string
	Error  string
	AuthId int
}

type App struct {
	// InfoLog      *log.Logger
	// ErrorLog     *log.Logger
	Session  *scs.SessionManager
	MailChan chan MailData
}

type Reservation struct {
	Id        int
	FirstName string
	LastName  string
	Email     string
	Phone     string
	StartDate time.Time
	EndDate   time.Time
	RoomId    int
	CreatedAt time.Time
	UpdatedAt time.Time
	Processed int
	Room      Room
}

type User struct {
	Id          int
	FirstName   string
	LastName    string
	Email       string
	Password    string
	AccessLevel int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Room struct {
	Id        int       `json:"id"`
	RoomName  string    `json:"room_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Slug      string    `json:"slug"`
}

type Restriction struct {
	Id              int
	RestrictionName string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type RoomRestriction struct {
	Id            int
	StartDate     time.Time
	EndDate       time.Time
	RoomId        int
	ReservationId int
	RestrictionId int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Room          Room
	Reservation   Reservation
	Restriction   Restriction
}

type MailData struct {
	To              string
	From            string
	Subject         string
	Content         string
	Template        string
	TemplateContent map[string]string
}
