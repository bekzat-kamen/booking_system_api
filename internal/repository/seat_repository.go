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
	ErrSeatNotFound     = errors.New("seat not found")
	ErrSeatNotAvailable = errors.New("seat is not available")
)

type SeatRepositoryInterface interface {
	Create(ctx context.Context, seat *model.Seat) error
	CreateBatch(ctx context.Context, seats []*model.Seat) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Seat, error)
	GetByEvent(ctx context.Context, eventID uuid.UUID) ([]*model.Seat, error)
	GetByEventAndStatus(ctx context.Context, eventID uuid.UUID, status model.SeatStatus) ([]*model.Seat, error)
	Update(ctx context.Context, seat *model.Seat) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.SeatStatus, version int) error
	CountByEvent(ctx context.Context, eventID uuid.UUID) (int, error)
	CountByStatus(ctx context.Context, eventID uuid.UUID, status model.SeatStatus) (int, error)
	Exists(ctx context.Context, eventID uuid.UUID, rowNumber, seatNumber string) (bool, error)
}

type SeatRepository struct {
	db *sqlx.DB
}

func NewSeatRepository(db *sqlx.DB) *SeatRepository {
	return &SeatRepository{db: db}
}

func (r *SeatRepository) Create(ctx context.Context, seat *model.Seat) error {
	query := `
		INSERT INTO seats (id, event_id, seat_number, row_number, status, price, version, created_at, updated_at)
		VALUES (:id, :event_id, :seat_number, :row_number, :status, :price, :version, :created_at, :updated_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":          seat.ID,
		"event_id":    seat.EventID,
		"seat_number": seat.SeatNumber,
		"row_number":  seat.RowNumber,
		"status":      seat.Status,
		"price":       seat.Price,
		"version":     seat.Version,
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
	})

	return err
}

func (r *SeatRepository) CreateBatch(ctx context.Context, seats []*model.Seat) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := `
		INSERT INTO seats (id, event_id, seat_number, row_number, status, price, version, created_at, updated_at)
		VALUES (:id, :event_id, :seat_number, :row_number, :status, :price, :version, :created_at, :updated_at)
	`

	for _, seat := range seats {
		_, err := tx.NamedExecContext(ctx, query, map[string]interface{}{
			"id":          seat.ID,
			"event_id":    seat.EventID,
			"seat_number": seat.SeatNumber,
			"row_number":  seat.RowNumber,
			"status":      seat.Status,
			"price":       seat.Price,
			"version":     seat.Version,
			"created_at":  time.Now(),
			"updated_at":  time.Now(),
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *SeatRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Seat, error) {
	query := `
		SELECT id, event_id, seat_number, row_number, status, price, version, created_at, updated_at
		FROM seats
		WHERE id = $1
	`

	seat := &model.Seat{}
	err := r.db.GetContext(ctx, seat, query, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSeatNotFound
		}
		return nil, err
	}

	return seat, nil
}

func (r *SeatRepository) GetByEvent(ctx context.Context, eventID uuid.UUID) ([]*model.Seat, error) {
	query := `
		SELECT id, event_id, seat_number, row_number, status, price, version, created_at, updated_at
		FROM seats
		WHERE event_id = $1
		ORDER BY row_number, seat_number
	`

	var seats []*model.Seat
	err := r.db.SelectContext(ctx, &seats, query, eventID)

	return seats, err
}

func (r *SeatRepository) GetByEventAndStatus(ctx context.Context, eventID uuid.UUID, status model.SeatStatus) ([]*model.Seat, error) {
	query := `
		SELECT id, event_id, seat_number, row_number, status, price, version, created_at, updated_at
		FROM seats
		WHERE event_id = $1 AND status = $2
		ORDER BY row_number, seat_number
	`

	var seats []*model.Seat
	err := r.db.SelectContext(ctx, &seats, query, eventID, status)

	return seats, err
}

func (r *SeatRepository) Update(ctx context.Context, seat *model.Seat) error {
	query := `
		UPDATE seats
		SET status = :status, price = :price, version = :version + 1, updated_at = :updated_at
		WHERE id = :id AND version = :version
	`

	result, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":         seat.ID,
		"status":     seat.Status,
		"price":      seat.Price,
		"version":    seat.Version,
		"updated_at": time.Now(),
	})
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("seat not found or version mismatch (optimistic lock)")
	}

	return nil
}

func (r *SeatRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.SeatStatus, version int) error {
	seat := &model.Seat{
		ID:      id,
		Status:  status,
		Version: version,
	}
	return r.Update(ctx, seat)
}

func (r *SeatRepository) CountByEvent(ctx context.Context, eventID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM seats WHERE event_id = $1`
	var count int
	err := r.db.GetContext(ctx, &count, query, eventID)
	return count, err
}

func (r *SeatRepository) CountByStatus(ctx context.Context, eventID uuid.UUID, status model.SeatStatus) (int, error) {
	query := `SELECT COUNT(*) FROM seats WHERE event_id = $1 AND status = $2`
	var count int
	err := r.db.GetContext(ctx, &count, query, eventID, status)
	return count, err
}

func (r *SeatRepository) Exists(ctx context.Context, eventID uuid.UUID, rowNumber, seatNumber string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM seats WHERE event_id = $1 AND row_number = $2 AND seat_number = $3)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, eventID, rowNumber, seatNumber).Scan(&exists)
	return exists, err
}
