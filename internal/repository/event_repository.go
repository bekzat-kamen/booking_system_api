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
	ErrEventNotFound = errors.New("event not found")
)

type EventRepositoryInterface interface {
	Create(ctx context.Context, event *model.Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Event, error)
	GetAll(ctx context.Context, limit, offset int) ([]*model.Event, error)
	Count(ctx context.Context) (int, error)
	GetByOrganizer(ctx context.Context, organizerID uuid.UUID, limit, offset int) ([]*model.Event, error)
	Update(ctx context.Context, event *model.Event) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type EventRepository struct {
	db *sqlx.DB
}

func NewEventRepository(db *sqlx.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(ctx context.Context, event *model.Event) error {
	query := `
		INSERT INTO events (id, title, description, venue, event_date, total_seats, available_seats, price, status, created_by, image_url, created_at, updated_at)
		VALUES (:id, :title, :description, :venue, :event_date, :total_seats, :available_seats, :price, :status, :created_by, :image_url, :created_at, :updated_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":              event.ID,
		"title":           event.Title,
		"description":     event.Description,
		"venue":           event.Venue,
		"event_date":      event.EventDate,
		"total_seats":     event.TotalSeats,
		"available_seats": event.AvailableSeats,
		"price":           event.Price,
		"status":          event.Status,
		"created_by":      event.CreatedBy,
		"image_url":       event.ImageURL,
		"created_at":      time.Now(),
		"updated_at":      time.Now(),
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *EventRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Event, error) {
	query := `
		SELECT id, title, description, venue, event_date, total_seats, available_seats, price, status, created_by, image_url, created_at, updated_at
		FROM events
		WHERE id = $1
	`

	event := &model.Event{}
	if err := r.db.GetContext(ctx, event, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEventNotFound
		}
		return nil, err
	}

	return event, nil
}

func (r *EventRepository) GetAll(ctx context.Context, limit, offset int) ([]*model.Event, error) {
	query := `
		SELECT id, title, description, venue, event_date, total_seats, available_seats, price, status, created_by, image_url, created_at, updated_at
		FROM events
		ORDER BY event_date ASC
		LIMIT $1 OFFSET $2
	`

	events := make([]*model.Event, 0)
	if err := r.db.SelectContext(ctx, &events, query, limit, offset); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *EventRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM events`

	var count int
	if err := r.db.GetContext(ctx, &count, query); err != nil {
		return 0, err
	}

	return count, nil
}

func (r *EventRepository) GetByOrganizer(ctx context.Context, organizerID uuid.UUID, limit, offset int) ([]*model.Event, error) {
	query := `
		SELECT id, title, description, venue, event_date, total_seats, available_seats, price, status, created_by, image_url, created_at, updated_at
		FROM events
		WHERE created_by = $1
		ORDER BY event_date ASC
		LIMIT $2 OFFSET $3
	`

	events := make([]*model.Event, 0)
	if err := r.db.SelectContext(ctx, &events, query, organizerID, limit, offset); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *EventRepository) Update(ctx context.Context, event *model.Event) error {
	query := `
		UPDATE events
		SET title = :title,
		    description = :description,
		    venue = :venue,
		    event_date = :event_date,
		    total_seats = :total_seats,
		    available_seats = :available_seats,
		    price = :price,
		    status = :status,
		    image_url = :image_url,
		    updated_at = :updated_at
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":              event.ID,
		"title":           event.Title,
		"description":     event.Description,
		"venue":           event.Venue,
		"event_date":      event.EventDate,
		"total_seats":     event.TotalSeats,
		"available_seats": event.AvailableSeats,
		"price":           event.Price,
		"status":          event.Status,
		"image_url":       event.ImageURL,
		"updated_at":      time.Now(),
	})
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

func (r *EventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM events WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
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
