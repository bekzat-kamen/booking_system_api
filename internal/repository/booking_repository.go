package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrBookingNotFound = errors.New("booking not found")
)

type BookingRepositoryInterface interface {
	Create(ctx context.Context, booking *model.Booking, seats []*model.BookingSeat) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Booking, error)
	GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Booking, error)
	GetByEvent(ctx context.Context, eventID uuid.UUID, limit, offset int) ([]*model.Booking, error)
	GetSeats(ctx context.Context, bookingID uuid.UUID) ([]*model.BookingSeat, error)
	Update(ctx context.Context, booking *model.Booking) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.BookingStatus) error
	CountByUser(ctx context.Context, userID uuid.UUID) (int, error)
	GetExpiredPending(ctx context.Context) ([]*model.Booking, error)
}

type BookingRepository struct {
	db *sqlx.DB
}

func NewBookingRepository(db *sqlx.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) Create(ctx context.Context, booking *model.Booking, seats []*model.BookingSeat) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO bookings (id, user_id, event_id, total_amount, discount, final_amount, status, expires_at, created_at, updated_at)
		VALUES (:id, :user_id, :event_id, :total_amount, :discount, :final_amount, :status, :expires_at, :created_at, :updated_at)
	`

	_, err = tx.NamedExecContext(ctx, query, map[string]interface{}{
		"id":           booking.ID,
		"user_id":      booking.UserID,
		"event_id":     booking.EventID,
		"total_amount": booking.TotalAmount,
		"discount":     booking.Discount,
		"final_amount": booking.FinalAmount,
		"status":       booking.Status,
		"expires_at":   booking.ExpiresAt,
		"created_at":   time.Now(),
		"updated_at":   time.Now(),
	})
	if err != nil {
		return err
	}

	seatQuery := `
		INSERT INTO booking_seats (booking_id, seat_id, price, created_at)
		VALUES (:booking_id, :seat_id, :price, :created_at)
	`

	for _, seat := range seats {
		_, err := tx.NamedExecContext(ctx, seatQuery, map[string]interface{}{
			"booking_id": seat.BookingID,
			"seat_id":    seat.SeatID,
			"price":      seat.Price,
			"created_at": time.Now(),
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *BookingRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	query := `
		SELECT id, user_id, event_id, total_amount, discount, final_amount, status, expires_at, paid_at, created_at, updated_at
		FROM bookings
		WHERE id = $1
	`

	booking := &model.Booking{}
	err := r.db.GetContext(ctx, booking, query, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}

	return booking, nil
}

func (r *BookingRepository) GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Booking, error) {
	query := `
		SELECT id, user_id, event_id, total_amount, discount, final_amount, status, expires_at, paid_at, created_at, updated_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var bookings []*model.Booking
	err := r.db.SelectContext(ctx, &bookings, query, userID, limit, offset)

	return bookings, err
}

func (r *BookingRepository) GetByEvent(ctx context.Context, eventID uuid.UUID, limit, offset int) ([]*model.Booking, error) {
	query := `
		SELECT id, user_id, event_id, total_amount, discount, final_amount, status, expires_at, paid_at, created_at, updated_at
		FROM bookings
		WHERE event_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var bookings []*model.Booking
	err := r.db.SelectContext(ctx, &bookings, query, eventID, limit, offset)

	return bookings, err
}

func (r *BookingRepository) GetSeats(ctx context.Context, bookingID uuid.UUID) ([]*model.BookingSeat, error) {
	query := `
		SELECT booking_id, seat_id, price, created_at
		FROM booking_seats
		WHERE booking_id = $1
	`

	var seats []*model.BookingSeat
	err := r.db.SelectContext(ctx, &seats, query, bookingID)

	return seats, err
}

func (r *BookingRepository) Update(ctx context.Context, booking *model.Booking) error {
	query := `
		UPDATE bookings
		SET status = :status, total_amount = :total_amount, discount = :discount, 
		    final_amount = :final_amount, expires_at = :expires_at, paid_at = :paid_at, updated_at = :updated_at
		WHERE id = :id
	`

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":           booking.ID,
		"status":       booking.Status,
		"total_amount": booking.TotalAmount,
		"discount":     booking.Discount,
		"final_amount": booking.FinalAmount,
		"expires_at":   booking.ExpiresAt,
		"paid_at":      booking.PaidAt,
		"updated_at":   time.Now(),
	})

	return err
}

func (r *BookingRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.BookingStatus) error {
	query := `UPDATE bookings SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	return err
}

func (r *BookingRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM bookings WHERE user_id = $1`
	var count int
	err := r.db.GetContext(ctx, &count, query, userID)
	return count, err
}

func (r *BookingRepository) GetExpiredPending(ctx context.Context) ([]*model.Booking, error) {
	query := `
		SELECT id, user_id, event_id, total_amount, discount, final_amount, status, expires_at, paid_at, created_at, updated_at
		FROM bookings
		WHERE status = 'pending' AND expires_at < NOW()
	`

	var bookings []*model.Booking
	err := r.db.SelectContext(ctx, &bookings, query)

	return bookings, err
}
