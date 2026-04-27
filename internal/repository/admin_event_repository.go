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

type AdminEventRepositoryInterface interface {
	GetAllEvents(ctx context.Context, page, limit int, status, organizerID string) ([]*model.Event, int, error)
	GetEventByID(ctx context.Context, id uuid.UUID) (*model.Event, error)
	UpdateEventAdmin(ctx context.Context, event *model.Event) error
	DeleteEventAdmin(ctx context.Context, id uuid.UUID) error
	PublishEventAdmin(ctx context.Context, id uuid.UUID) error
	GetEventStats(ctx context.Context, eventID uuid.UUID) (map[string]interface{}, error)
	GetEventsByStatus(ctx context.Context) (map[string]int64, error)
}

type AdminEventRepository struct {
	db *sqlx.DB
}

func NewAdminEventRepository(db *sqlx.DB) *AdminEventRepository {
	return &AdminEventRepository{db: db}
}

func (r *AdminEventRepository) GetAllEvents(ctx context.Context, page, limit int, status, organizerID string) ([]*model.Event, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := `
		SELECT e.id, e.title, e.description, e.venue, e.event_date, e.total_seats, 
		       e.available_seats, e.price, e.status, e.created_by, e.image_url, 
		       e.created_at, e.updated_at, u.email as organizer_email
		FROM events e
		LEFT JOIN users u ON e.created_by = u.id
		WHERE ($1 = '' OR e.status = $1)
		  AND ($2 = '' OR e.created_by::text = $2)
		ORDER BY e.created_at DESC
		LIMIT $3 OFFSET $4
	`

	var events []*model.Event
	err := r.db.SelectContext(ctx, &events, query, status, organizerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	countQuery := `
		SELECT COUNT(*) FROM events
		WHERE ($1 = '' OR status = $1)
		  AND ($2 = '' OR created_by::text = $2)
	`
	var total int
	err = r.db.GetContext(ctx, &total, countQuery, status, organizerID)
	if err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (r *AdminEventRepository) GetEventByID(ctx context.Context, id uuid.UUID) (*model.Event, error) {
	query := `
		SELECT e.id, e.title, e.description, e.venue, e.event_date, e.total_seats,
		       e.available_seats, e.price, e.status, e.created_by, e.image_url,
		       e.created_at, e.updated_at, u.email as organizer_email
		FROM events e
		LEFT JOIN users u ON e.created_by = u.id
		WHERE e.id = $1
	`

	event := &model.Event{}
	err := r.db.GetContext(ctx, event, query, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEventNotFound
		}
		return nil, err
	}

	return event, nil
}

func (r *AdminEventRepository) UpdateEventAdmin(ctx context.Context, event *model.Event) error {
	query := `
		UPDATE events
		SET title = $1, description = $2, venue = $3, event_date = $4,
		    total_seats = $5, available_seats = $6, price = $7, status = $8,
		    image_url = $9, updated_at = $10
		WHERE id = $11
	`

	result, err := r.db.ExecContext(ctx, query,
		event.Title,
		event.Description,
		event.Venue,
		event.EventDate,
		event.TotalSeats,
		event.AvailableSeats,
		event.Price,
		event.Status,
		event.ImageURL,
		time.Now(),
		event.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrEventNotFound
	}

	return nil
}

func (r *AdminEventRepository) DeleteEventAdmin(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE events SET status = 'cancelled', updated_at = $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrEventNotFound
	}

	return nil
}

func (r *AdminEventRepository) PublishEventAdmin(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE events SET status = 'published', updated_at = $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrEventNotFound
	}

	return nil
}
func (r *AdminEventRepository) GetEventStats(ctx context.Context, eventID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var bookings int64
	if err := r.db.GetContext(ctx, &bookings, `SELECT COUNT(*) FROM bookings WHERE event_id = $1`, eventID); err != nil {
		return nil, err
	}
	stats["total_bookings"] = bookings

	var confirmed int64
	if err := r.db.GetContext(ctx, &confirmed, `SELECT COUNT(*) FROM bookings WHERE event_id = $1 AND status = 'confirmed'`, eventID); err != nil {
		return nil, err
	}
	stats["confirmed_bookings"] = confirmed

	var revenue float64
	if err := r.db.GetContext(ctx, &revenue, `SELECT COALESCE(SUM(final_amount), 0) FROM bookings WHERE event_id = $1 AND status = 'confirmed'`, eventID); err != nil {
		return nil, err
	}
	stats["total_revenue"] = revenue

	var totalSeats, availableSeats int
	if err := r.db.GetContext(ctx, &totalSeats, `SELECT COALESCE(SUM(total_seats), 0) FROM events WHERE id = $1`, eventID); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &availableSeats, `SELECT COALESCE(SUM(available_seats), 0) FROM events WHERE id = $1`, eventID); err != nil {
		return nil, err
	}
	stats["total_seats"] = totalSeats
	stats["sold_seats"] = totalSeats - availableSeats

	return stats, nil
}

func (r *AdminEventRepository) GetEventsByStatus(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	queries := map[string]string{
		"total":     `SELECT COUNT(*) FROM events`,
		"draft":     `SELECT COUNT(*) FROM events WHERE status = 'draft'`,
		"published": `SELECT COUNT(*) FROM events WHERE status = 'published'`,
		"sold_out":  `SELECT COUNT(*) FROM events WHERE status = 'sold_out'`,
		"cancelled": `SELECT COUNT(*) FROM events WHERE status = 'cancelled'`,
		"completed": `SELECT COUNT(*) FROM events WHERE status = 'completed'`,
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
