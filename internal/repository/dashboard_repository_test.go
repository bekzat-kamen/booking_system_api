package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDashboardMock(t *testing.T) (*DashboardRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewDashboardRepository(sqlxDB)

	return repo, mock, func() {
		_ = db.Close()
	}
}

func TestDashboardRepository_GetStats(t *testing.T) {
	repo, mock, cleanup := setupDashboardMock(t)
	defer cleanup()

	mock.ExpectQuery("SELECT COUNT(*) FROM users").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))
	mock.ExpectQuery("SELECT COUNT(*) FROM users WHERE created_at >= NOW() - INTERVAL '1 day'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery("SELECT COUNT(*) FROM users WHERE status = 'active'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(80))

	mock.ExpectQuery("SELECT COUNT(*) FROM events").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(20))
	mock.ExpectQuery("SELECT COUNT(*) FROM events WHERE status = 'published'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(15))
	mock.ExpectQuery("SELECT COUNT(*) FROM events WHERE status = 'draft'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery("SELECT COUNT(*) FROM events WHERE status = 'cancelled'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery("SELECT COUNT(*) FROM bookings").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(500))
	mock.ExpectQuery("SELECT COUNT(*) FROM bookings WHERE status = 'pending'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(50))
	mock.ExpectQuery("SELECT COUNT(*) FROM bookings WHERE status = 'confirmed'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(400))
	mock.ExpectQuery("SELECT COUNT(*) FROM bookings WHERE status = 'cancelled'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(50))

	mock.ExpectQuery("SELECT COALESCE(SUM(final_amount), 0) FROM bookings WHERE status = 'confirmed'").WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(10000.0))
	mock.ExpectQuery("SELECT COALESCE(SUM(final_amount), 0) FROM bookings WHERE status = 'confirmed' AND paid_at >= NOW() - INTERVAL '1 day'").WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(500.0))
	mock.ExpectQuery("SELECT COALESCE(SUM(final_amount), 0) FROM bookings WHERE status = 'confirmed' AND paid_at >= NOW() - INTERVAL '30 day'").WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(5000.0))

	mock.ExpectQuery("SELECT COALESCE(AVG(final_amount), 0) FROM bookings WHERE status = 'confirmed'").WillReturnRows(sqlmock.NewRows([]string{"avg"}).AddRow(25.0))

	mock.ExpectQuery("SELECT COUNT(*) FROM promocodes").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	mock.ExpectQuery("SELECT COUNT(*) FROM promocodes WHERE is_active = true").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))

	mock.ExpectQuery("SELECT COUNT(*) FROM payment_transactions").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(500))
	mock.ExpectQuery("SELECT COUNT(*) FROM payment_transactions WHERE status = 'success'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(450))
	mock.ExpectQuery("SELECT COUNT(*) FROM payment_transactions WHERE status = 'failed'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(40))
	mock.ExpectQuery("SELECT COUNT(*) FROM payment_transactions WHERE status = 'refunded'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	stats, err := repo.GetStats(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(100), stats.TotalUsers)
	assert.Equal(t, 10000.0, stats.TotalRevenue)
}
