package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupEventMock(t *testing.T) (*EventRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewEventRepository(sqlxDB)

	return repo, mock, func() {
		_ = db.Close()
	}
}

func TestEventRepository_Create(t *testing.T) {
	repo, mock, cleanup := setupEventMock(t)
	defer cleanup()

	event := &model.Event{
		ID:             uuid.New(),
		Title:          "Concert",
		Venue:          "Hall",
		EventDate:      time.Now(),
		TotalSeats:     100,
		AvailableSeats: 100,
		Price:          50.0,
		Status:         model.EventStatusDraft,
		CreatedBy:      uuid.New(),
	}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO events")).
		WithArgs(
			event.ID,
			event.Title,
			event.Description,
			event.Venue,
			event.EventDate,
			event.TotalSeats,
			event.AvailableSeats,
			event.Price,
			event.Status,
			event.CreatedBy,
			event.ImageURL,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), event)
	assert.NoError(t, err)
}

func TestEventRepository_GetByID(t *testing.T) {
	repo, mock, cleanup := setupEventMock(t)
	defer cleanup()

	eventID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "title", "description", "venue", "event_date", "total_seats", "available_seats", "price", "status", "created_by", "image_url", "created_at", "updated_at"}).
			AddRow(eventID, "Title", "Desc", "Venue", time.Now(), 100, 100, 50.0, "published", uuid.New(), "", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title, description, venue, event_date, total_seats, available_seats, price, status, created_by, image_url, created_at, updated_at FROM events WHERE id = $1")).
			WithArgs(eventID).
			WillReturnRows(rows)

		event, err := repo.GetByID(context.Background(), eventID)
		require.NoError(t, err)
		assert.Equal(t, eventID, event.ID)
	})

	t.Run("Not Found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title, description, venue, event_date, total_seats, available_seats, price, status, created_by, image_url, created_at, updated_at FROM events WHERE id = $1")).
			WithArgs(eventID).
			WillReturnError(sql.ErrNoRows)

		event, err := repo.GetByID(context.Background(), eventID)
		assert.ErrorIs(t, err, ErrEventNotFound)
		assert.Nil(t, event)
	})
}

func TestEventRepository_GetAll(t *testing.T) {
	repo, mock, cleanup := setupEventMock(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title, description, venue, event_date, total_seats, available_seats, price, status, created_by, image_url, created_at, updated_at FROM events ORDER BY event_date ASC LIMIT $1 OFFSET $2")).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow(uuid.New(), "E1").AddRow(uuid.New(), "E2"))

	events, err := repo.GetAll(context.Background(), 10, 0)
	assert.NoError(t, err)
	assert.Len(t, events, 2)
}

func TestEventRepository_Update(t *testing.T) {
	repo, mock, cleanup := setupEventMock(t)
	defer cleanup()

	event := &model.Event{
		ID:    uuid.New(),
		Title: "New Title",
	}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta("UPDATE events")).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(context.Background(), event)
		assert.NoError(t, err)
	})

	t.Run("Not Found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta("UPDATE events")).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(context.Background(), event)
		assert.ErrorIs(t, err, ErrEventNotFound)
	})
}

func TestEventRepository_Delete(t *testing.T) {
	repo, mock, cleanup := setupEventMock(t)
	defer cleanup()

	eventID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM events WHERE id = $1")).
			WithArgs(eventID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(context.Background(), eventID)
		assert.NoError(t, err)
	})

	t.Run("Not Found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM events WHERE id = $1")).
			WithArgs(eventID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(context.Background(), eventID)
		assert.ErrorIs(t, err, ErrEventNotFound)
	})
}
