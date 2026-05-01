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

func setupAdminBookingMock(t *testing.T) (*AdminBookingRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewAdminBookingRepository(sqlxDB)

	return repo, mock, func() {
		_ = db.Close()
	}
}

func TestAdminBookingRepository_GetAllBookings(t *testing.T) {
	repo, mock, cleanup := setupAdminBookingMock(t)
	defer cleanup()

	t.Run("Success", func(t *testing.T) {
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
		mock.ExpectQuery(query).
			WithArgs("confirmed", "", "", 20, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_amount"}).AddRow(uuid.New(), 100))

		countQuery := `
		SELECT COUNT(*) FROM bookings
		WHERE ($1 = '' OR status = $1)
		  AND ($2 = '' OR user_id::text = $2)
		  AND ($3 = '' OR event_id::text = $3)
	`
		mock.ExpectQuery(countQuery).
			WithArgs("confirmed", "", "").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		bookings, total, err := repo.GetAllBookings(context.Background(), 1, 20, "confirmed", "", "")
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, bookings, 1)
	})
}

func TestAdminBookingRepository_GetBookingsStats(t *testing.T) {
	repo, mock, cleanup := setupAdminBookingMock(t)
	defer cleanup()

	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery("SELECT COUNT(*) FROM bookings").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))
	mock.ExpectQuery("SELECT COUNT(*) FROM bookings WHERE status = 'pending'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	mock.ExpectQuery("SELECT COUNT(*) FROM bookings WHERE status = 'confirmed'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(70))
	mock.ExpectQuery("SELECT COUNT(*) FROM bookings WHERE status = 'cancelled'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery("SELECT COUNT(*) FROM bookings WHERE status = 'expired'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery("SELECT COUNT(*) FROM bookings WHERE status = 'completed'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	stats, err := repo.GetBookingsStats(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(100), stats["total"])
	assert.Equal(t, int64(70), stats["confirmed"])
}
