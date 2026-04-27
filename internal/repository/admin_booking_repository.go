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

type AdminBookingRepositoryInterface interface {
	GetAllBookings(ctx context.Context, page, limit int, status, userID, eventID string) ([]*model.Booking, int, error)
	GetBookingByID(ctx context.Context, id uuid.UUID) (*model.Booking, error)
	GetBookingSeats(ctx context.Context, bookingID uuid.UUID) ([]*model.BookingSeat, error)
	GetBookingPayment(ctx context.Context, bookingID uuid.UUID) (*model.Payment, error)
	CancelBookingAdmin(ctx context.Context, id uuid.UUID, status model.BookingStatus) error
	RefundPayment(ctx context.Context, paymentID uuid.UUID) error
	GetBookingsStats(ctx context.Context) (map[string]int64, error)
	GetRevenueStats(ctx context.Context) (map[string]float64, error)
}

type AdminBookingRepository struct {
	db *sqlx.DB
}

func NewAdminBookingRepository(db *sqlx.DB) *AdminBookingRepository {
	return &AdminBookingRepository{db: db}
}

func (r *AdminBookingRepository) GetAllBookings(ctx context.Context, page, limit int, status, userID, eventID string) ([]*model.Booking, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := `
		SELECT b.id, b.user_id, b.event_id, b.total_amount, b.discount, b.final_amount,
		       b.status, b.expires_at, b.paid_at, b.created_at, b.updated_at,
		       u.email as user_email, e.title as event_title
		FROM bookings b
		LEFT JOIN users u ON b.user_id = u.id
		LEFT JOIN events e ON b.event_id = e.id
		WHERE ($1 = '' OR b.status = $1)
		  AND ($2 = '' OR b.user_id::text = $2)
		  AND ($3 = '' OR b.event_id::text = $3)
		ORDER BY b.created_at DESC
		LIMIT $4 OFFSET $5
	`

	var bookings []*model.Booking
	err := r.db.SelectContext(ctx, &bookings, query, status, userID, eventID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	countQuery := `
		SELECT COUNT(*) FROM bookings
		WHERE ($1 = '' OR status = $1)
		  AND ($2 = '' OR user_id::text = $2)
		  AND ($3 = '' OR event_id::text = $3)
	`
	var total int
	err = r.db.GetContext(ctx, &total, countQuery, status, userID, eventID)
	if err != nil {
		return nil, 0, err
	}

	return bookings, total, nil
}

func (r *AdminBookingRepository) GetBookingByID(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	query := `
		SELECT b.id, b.user_id, b.event_id, b.total_amount, b.discount, b.final_amount,
		       b.status, b.expires_at, b.paid_at, b.created_at, b.updated_at,
		       u.email as user_email, u.name as user_name, e.title as event_title,
		       e.event_date as event_date, e.venue as event_venue
		FROM bookings b
		LEFT JOIN users u ON b.user_id = u.id
		LEFT JOIN events e ON b.event_id = e.id
		WHERE b.id = $1
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

func (r *AdminBookingRepository) GetBookingSeats(ctx context.Context, bookingID uuid.UUID) ([]*model.BookingSeat, error) {
	query := `
		SELECT bs.booking_id, bs.seat_id, bs.price, bs.created_at,
		       s.seat_number, s.row_number, s.status as seat_status
		FROM booking_seats bs
		LEFT JOIN seats s ON bs.seat_id = s.id
		WHERE bs.booking_id = $1
	`

	var seats []*model.BookingSeat
	err := r.db.SelectContext(ctx, &seats, query, bookingID)
	return seats, err
}

func (r *AdminBookingRepository) GetBookingPayment(ctx context.Context, bookingID uuid.UUID) (*model.Payment, error) {
	query := `
		SELECT id, booking_id, transaction_id, amount, status, payment_method,
		       provider, provider_response, paid_at, created_at, updated_at
		FROM payment_transactions
		WHERE booking_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	payment := &model.Payment{}
	err := r.db.GetContext(ctx, payment, query, bookingID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	return payment, nil
}

func (r *AdminBookingRepository) CancelBookingAdmin(ctx context.Context, id uuid.UUID, status model.BookingStatus) error {
	query := `UPDATE bookings SET status = $1, updated_at = $2 WHERE id = $3`
	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrBookingNotFound
	}

	return nil
}

func (r *AdminBookingRepository) RefundPayment(ctx context.Context, paymentID uuid.UUID) error {
	query := `UPDATE payment_transactions SET status = 'refunded', paid_at = NULL, updated_at = $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, time.Now(), paymentID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrPaymentNotFound
	}

	return nil
}

func (r *AdminBookingRepository) GetBookingsStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	queries := map[string]string{
		"total":     `SELECT COUNT(*) FROM bookings`,
		"pending":   `SELECT COUNT(*) FROM bookings WHERE status = 'pending'`,
		"confirmed": `SELECT COUNT(*) FROM bookings WHERE status = 'confirmed'`,
		"cancelled": `SELECT COUNT(*) FROM bookings WHERE status = 'cancelled'`,
		"expired":   `SELECT COUNT(*) FROM bookings WHERE status = 'expired'`,
		"completed": `SELECT COUNT(*) FROM bookings WHERE status = 'completed'`,
	}

	for key, query := range queries {
		var value int64
		if err := r.db.GetContext(ctx, &value, query); err != nil {
			return nil, err
		}
		stats[key] = value
	}

	return stats, nil
}

func (r *AdminBookingRepository) GetRevenueStats(ctx context.Context) (map[string]float64, error) {
	stats := make(map[string]float64)

	queries := map[string]string{
		"total_revenue":   `SELECT COALESCE(SUM(final_amount), 0) FROM bookings WHERE status = 'confirmed'`,
		"today_revenue":   `SELECT COALESCE(SUM(final_amount), 0) FROM bookings WHERE status = 'confirmed' AND paid_at >= NOW() - INTERVAL '1 day'`,
		"month_revenue":   `SELECT COALESCE(SUM(final_amount), 0) FROM bookings WHERE status = 'confirmed' AND paid_at >= NOW() - INTERVAL '30 day'`,
		"refunded_amount": `SELECT COALESCE(SUM(amount), 0) FROM payment_transactions WHERE status = 'refunded'`,
	}

	for key, query := range queries {
		var value float64
		if err := r.db.GetContext(ctx, &value, query); err != nil {
			return nil, err
		}
		stats[key] = value
	}

	return stats, nil
}
