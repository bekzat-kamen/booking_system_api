package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAdminEventMock(t *testing.T) (*AdminEventRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewAdminEventRepository(sqlxDB)

	return repo, mock, func() {
		_ = db.Close()
	}
}

func TestAdminEventRepository_GetAllEvents(t *testing.T) {
	repo, mock, cleanup := setupAdminEventMock(t)
	defer cleanup()

	t.Run("Success", func(t *testing.T) {
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
		mock.ExpectQuery(query).
			WithArgs("published", "", 20, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow(uuid.New(), "Concert"))

		countQuery := `
		SELECT COUNT(*) FROM events
		WHERE ($1 = '' OR status = $1)
		  AND ($2 = '' OR created_by::text = $2)
	`
		mock.ExpectQuery(countQuery).
			WithArgs("published", "").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		events, total, err := repo.GetAllEvents(context.Background(), 1, 20, "published", "")
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, events, 1)
	})
}

func TestAdminEventRepository_GetEventStats(t *testing.T) {
	repo, mock, cleanup := setupAdminEventMock(t)
	defer cleanup()

	eventID := uuid.New()

	mock.ExpectQuery("SELECT COUNT(*) FROM bookings WHERE event_id = $1").WithArgs(eventID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	mock.ExpectQuery("SELECT COUNT(*) FROM bookings WHERE event_id = $1 AND status = 'confirmed'").WithArgs(eventID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(8))
	mock.ExpectQuery("SELECT COALESCE(SUM(final_amount), 0) FROM bookings WHERE event_id = $1 AND status = 'confirmed'").WithArgs(eventID).WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(400.0))
	mock.ExpectQuery("SELECT COALESCE(SUM(total_seats), 0) FROM events WHERE id = $1").WithArgs(eventID).WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(100))
	mock.ExpectQuery("SELECT COALESCE(SUM(available_seats), 0) FROM events WHERE id = $1").WithArgs(eventID).WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(92))

	stats, err := repo.GetEventStats(context.Background(), eventID)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), stats["total_bookings"])
	assert.Equal(t, int64(8), stats["confirmed_bookings"])
	assert.Equal(t, 400.0, stats["total_revenue"])
}

func TestAdminEventRepository_GetEventByID(t *testing.T) {
	repo, mock, cleanup := setupAdminEventMock(t)
	defer cleanup()

	eventID := uuid.New()

	query := `
		SELECT e.id, e.title, e.description, e.venue, e.event_date, e.total_seats, 
		       e.available_seats, e.price, e.status, e.created_by, e.image_url, 
		       e.created_at, e.updated_at, u.email as organizer_email
		FROM events e
		LEFT JOIN users u ON e.created_by = u.id
		WHERE e.id = $1
	`
	mock.ExpectQuery(query).WithArgs(eventID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(eventID))

	event, err := repo.GetEventByID(context.Background(), eventID)
	assert.NoError(t, err)
	assert.Equal(t, eventID, event.ID)
}

func TestAdminEventRepository_PublishEventAdmin(t *testing.T) {
	repo, mock, cleanup := setupAdminEventMock(t)
	defer cleanup()

	eventID := uuid.New()

	mock.ExpectExec("UPDATE events SET status = 'published', updated_at = $1 WHERE id = $2").
		WithArgs(sqlmock.AnyArg(), eventID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.PublishEventAdmin(context.Background(), eventID)
	assert.NoError(t, err)
}

func TestAdminEventRepository_GetEventsByStatus(t *testing.T) {
	repo, mock, cleanup := setupAdminEventMock(t)
	defer cleanup()

	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery("SELECT COUNT(*) FROM events").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(20))
	mock.ExpectQuery("SELECT COUNT(*) FROM events WHERE status = 'draft'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery("SELECT COUNT(*) FROM events WHERE status = 'published'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(15))
	mock.ExpectQuery("SELECT COUNT(*) FROM events WHERE status = 'sold_out'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT COUNT(*) FROM events WHERE status = 'cancelled'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT COUNT(*) FROM events WHERE status = 'completed'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	stats, err := repo.GetEventsByStatus(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(20), stats["total"])
}

