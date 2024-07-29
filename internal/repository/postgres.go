package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Noblefel/InnOne-bookings-web-app/internal/types"
	"golang.org/x/crypto/bcrypt"
)

type postgresDBRepo struct {
	app *types.App
	db  *sql.DB
}

func NewPostgresRepo(conn *sql.DB, a *types.App) *postgresDBRepo {
	return &postgresDBRepo{a, conn}
}

func (repo *postgresDBRepo) GetReservationById(id int) (types.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	SELECT r.*, rm.id, rm.room_name, rm.slug
	FROM reservations r
	LEFT JOIN rooms rm ON (r.room_id = rm.id)
	WHERE r.id = $1
	`

	var res types.Reservation

	err := repo.db.QueryRowContext(ctx, query, id).Scan(
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

func (repo *postgresDBRepo) GetNewReservations() ([]types.Reservation, error) {
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

	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var reservations []types.Reservation

	for rows.Next() {
		var res types.Reservation

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

func (repo *postgresDBRepo) GetAllReservations() ([]types.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	SELECT 
		r.*, rm.id, rm.room_name, rm.slug
	FROM reservations r
	LEFT JOIN rooms rm ON (r.room_id = rm.id)
	ORDER BY r.start_date ASC
	`

	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservations []types.Reservation

	for rows.Next() {
		var res types.Reservation

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

func (repo *postgresDBRepo) UpdateReservation(r types.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	UPDATE reservations 
	SET first_name = $1, last_name = $2, email = $3, phone = $4, updated_at = $5
	WHERE id = $6
	`

	_, err := repo.db.ExecContext(ctx, query,
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

func (repo *postgresDBRepo) DeleteReservation(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `DELETE from reservations WHERE id = $1`

	_, err := repo.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (repo *postgresDBRepo) ApproveReservation(id, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	UPDATE reservations 
	SET processed = $1
	WHERE id = $2
	`

	_, err := repo.db.ExecContext(ctx, query, processed, id)
	if err != nil {
		return err
	}

	return nil
}

func (repo *postgresDBRepo) InsertReservation(r types.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	INSERT into reservations (first_name, last_name, email, phone, start_date, end_date, room_id, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id
	`
	var newId int

	err := repo.db.QueryRowContext(ctx, query,
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

func (repo *postgresDBRepo) InsertRoomRestriction(rr types.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	INSERT into room_restrictions (start_date, end_date, room_id, reservation_id, created_at, updated_at, restriction_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := repo.db.ExecContext(ctx, query,
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

func (repo *postgresDBRepo) SearchRoomAvailabilityByDates(start, end time.Time, roomId int) (bool, error) {
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

	err := repo.db.QueryRowContext(ctx, query, roomId, start, end).Scan(&numRows)
	if err != nil {
		return false, err
	}

	if numRows != 0 {
		return false, nil
	}

	return true, nil
}

func (repo *postgresDBRepo) SearchAllRoomsAvailibility(start, end time.Time) ([]types.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT	r.id, r.room_name, r.slug
		FROM rooms r
		WHERE r.id not in
		(select room_id from room_restrictions rr where $1 < rr.end_date and $2 > rr.start_date)
	`

	rows, err := repo.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, err
	}

	var rooms []types.Room

	for rows.Next() {
		var room types.Room
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

func (repo *postgresDBRepo) GetRoomBySlug(slug string) (types.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room types.Room

	query := `SELECT id, room_name, slug FROM rooms WHERE slug = $1`

	err := repo.db.QueryRowContext(ctx, query, slug).Scan(&room.Id, &room.RoomName, &room.Slug)
	if err != nil && err != sql.ErrNoRows {
		return room, err
	}

	return room, nil
}

func (repo *postgresDBRepo) GetAllRooms() ([]types.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT id, room_name, slug FROM rooms ORDER BY id"

	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var rooms []types.Room

	for rows.Next() {
		var room types.Room

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

func (repo *postgresDBRepo) GetUserById(id int) (types.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, first_name, last_name, email, password, access_level, created_at, updated_at FROM users WHERE id = $1`

	var user types.User
	err := repo.db.QueryRowContext(ctx, query, id).Scan(
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

func (repo *postgresDBRepo) UpdateUser(user types.User) error {
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

	_, err := repo.db.ExecContext(ctx, query,
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

func (repo *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	query := "SELECT id, password FROM users WHERE email = $1"

	err := repo.db.QueryRowContext(ctx, query, email).Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil
}
