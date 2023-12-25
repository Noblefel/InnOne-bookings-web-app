package dbrepo

import (
	"context"
	"database/sql"
	"time"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

func (m *postgresDBRepo) GetReservationById(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	SELECT r.*, rm.id, rm.room_name, rm.slug
	FROM reservations r
	LEFT JOIN rooms rm ON (r.room_id = rm.id)
	WHERE r.id = $1
	`

	var res models.Reservation

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&res.Id,
		&res.FirstName,
		&res.LastName,
		&res.Email,
		&res.Phone,
		&res.StartDate,
		&res.EndDate,
		&res.RoomId,
		&res.CreatedAt,
		&res.UpdatedAt,
		&res.Processed,
		&res.Room.Id,
		&res.Room.RoomName,
		&res.Room.Slug,
	)

	if err != nil {
		return res, err
	}

	return res, nil
}

func (m *postgresDBRepo) GetNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	SELECT 
		r.*, rm.id, rm.room_name, rm.slug
	FROM reservations r
	LEFT JOIN rooms rm ON (r.room_id = rm.id)
	WHERE processed = 0
	ORDER BY r.start_date ASC 
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var reservations []models.Reservation

	for rows.Next() {
		var res models.Reservation

		err := rows.Scan(
			&res.Id,
			&res.FirstName,
			&res.LastName,
			&res.Email,
			&res.Phone,
			&res.StartDate,
			&res.EndDate,
			&res.RoomId,
			&res.CreatedAt,
			&res.UpdatedAt,
			&res.Processed,
			&res.Room.Id,
			&res.Room.RoomName,
			&res.Room.Slug,
		)

		if err != nil {
			return nil, err
		}

		reservations = append(reservations, res)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reservations, nil
}

func (m *postgresDBRepo) GetAllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	SELECT 
		r.*, rm.id, rm.room_name, rm.slug
	FROM reservations r
	LEFT JOIN rooms rm ON (r.room_id = rm.id)
	ORDER BY r.start_date ASC
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservations []models.Reservation

	for rows.Next() {
		var res models.Reservation

		err := rows.Scan(
			&res.Id,
			&res.FirstName,
			&res.LastName,
			&res.Email,
			&res.Phone,
			&res.StartDate,
			&res.EndDate,
			&res.RoomId,
			&res.CreatedAt,
			&res.UpdatedAt,
			&res.Processed,
			&res.Room.Id,
			&res.Room.RoomName,
			&res.Room.Slug,
		)

		if err != nil {
			return nil, err
		}

		reservations = append(reservations, res)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reservations, nil
}

func (m *postgresDBRepo) UpdateReservation(r models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	UPDATE reservations 
	SET first_name = $1, last_name = $2, email = $3, phone = $4, updated_at = $5
	WHERE id = $6
	`

	_, err := m.DB.ExecContext(ctx, query,
		r.FirstName,
		r.LastName,
		r.Email,
		r.Phone,
		time.Now(),
		r.Id,
	)

	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) DeleteReservation(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `DELETE from reservations WHERE id = $1`

	_, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) ApproveReservation(id, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	UPDATE reservations 
	SET processed = $1
	WHERE id = $2
	`

	_, err := m.DB.ExecContext(ctx, query, processed, id)
	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) InsertReservation(r models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	INSERT into reservations (first_name, last_name, email, phone, start_date, end_date, room_id, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id
	`
	var newId int

	err := m.DB.QueryRowContext(ctx, query,
		r.FirstName,
		r.LastName,
		r.Email,
		r.Phone,
		r.StartDate,
		r.EndDate,
		r.RoomId,
		time.Now(),
		time.Now(),
	).Scan(&newId)

	if err != nil {
		return 0, err
	}

	return newId, nil
}

func (m *postgresDBRepo) InsertRoomRestriction(rr models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	INSERT into room_restrictions (start_date, end_date, room_id, reservation_id, created_at, updated_at, restriction_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := m.DB.ExecContext(ctx, query,
		rr.StartDate,
		rr.EndDate,
		rr.RoomId,
		rr.ReservationId,
		time.Now(),
		time.Now(),
		rr.RestrictionId,
	)

	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) SearchRoomAvailabilityByDates(start, end time.Time, roomId int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT count(id)
		FROM room_restrictions
		WHERE 
			room_id = $1 
			and $2 < end_date and $3 > start_date
	`

	var numRows int

	err := m.DB.QueryRowContext(ctx, query, roomId, start, end).Scan(&numRows)
	if err != nil {
		return false, err
	}

	if numRows != 0 {
		return false, nil
	}

	return true, nil
}

func (m *postgresDBRepo) SearchAllRoomsAvailibility(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT	r.id, r.room_name, r.slug
		FROM rooms r
		WHERE r.id not in
		(select room_id from room_restrictions rr where $1 < rr.end_date and $2 > rr.start_date)
	`

	rows, err := m.DB.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, err
	}

	var rooms []models.Room

	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.Id, &room.RoomName, &room.Slug)
		if err != nil {
			return nil, err
		}

		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rooms, nil
}

func (m *postgresDBRepo) GetRoomBySlug(slug string) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	query := `SELECT id, room_name, slug FROM rooms WHERE slug = $1`

	err := m.DB.QueryRowContext(ctx, query, slug).Scan(&room.Id, &room.RoomName, &room.Slug)
	if err != nil && err != sql.ErrNoRows {
		return room, err
	}

	return room, nil
}

func (m *postgresDBRepo) GetAllRooms() ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT id, room_name, slug FROM rooms ORDER BY id"

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var rooms []models.Room

	for rows.Next() {
		var room models.Room

		err := rows.Scan(&room.Id, &room.RoomName, &room.Slug)
		if err != nil {
			return nil, err
		}

		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rooms, nil
}

func (m *postgresDBRepo) GetUserById(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, first_name, last_name, email, password, access_level, created_at, updated_at FROM users WHERE id = $1`

	var user models.User
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.AccessLevel,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil && err != sql.ErrNoRows {
		return user, err
	}

	return user, nil
}

func (m *postgresDBRepo) UpdateUser(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		UPDATE users 
		SET 
			first_name = $1, 
			last_name = $2, 
			email = $3,
			access_level = $4,
			updated_at = $5   
		WHERE id = $1
	`

	_, err := m.DB.ExecContext(ctx, query,
		user.FirstName,
		user.LastName,
		user.Email,
		user.AccessLevel,
		time.Now(),
	)

	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	query := "SELECT id, password FROM users WHERE email = $1"

	err := m.DB.QueryRowContext(ctx, query, email).Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil
}
