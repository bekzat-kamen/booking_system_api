package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
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
		mock.ExpectClose()
		err := db.Close()
		require.NoError(t, err)
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

func TestDashboardRepository_GetStats_Error(t *testing.T) {
	repo, mock, cleanup := setupDashboardMock(t)
	defer cleanup()

	mock.ExpectQuery("SELECT COUNT(*) FROM users").WillReturnError(context.DeadlineExceeded)

	stats, err := repo.GetStats(context.Background())
	assert.Error(t, err)
	assert.Nil(t, stats)
}

func TestDashboardRepository_GetRevenueByDay(t *testing.T) {
	repo, mock, cleanup := setupDashboardMock(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"date", "amount"}).
		AddRow("2023-01-01", 100.0).
		AddRow("2023-01-02", 200.0)

	query := `
		SELECT 
			TO_CHAR(paid_at, 'YYYY-MM-DD') as date,
			COALESCE(SUM(final_amount), 0) as amount
		FROM bookings
		WHERE status = 'confirmed' 
		  AND paid_at >= NOW() - INTERVAL '30 day'
		GROUP BY TO_CHAR(paid_at, 'YYYY-MM-DD')
		ORDER BY date ASC
	`
	mock.ExpectQuery(query).WillReturnRows(rows)

	result, err := repo.GetRevenueByDay(context.Background())
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "2023-01-01", result[0].Date)
}

func TestDashboardRepository_GetTopEvents(t *testing.T) {
	repo, mock, cleanup := setupDashboardMock(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"event_id", "title", "bookings", "revenue", "event_date"}).
		AddRow(uuid.New(), "Event 1", 10, 1000.0, time.Now())

	query := `
		SELECT 
			e.id as event_id,
			e.title,
			COUNT(b.id) as bookings,
			COALESCE(SUM(b.final_amount), 0) as revenue,
			e.event_date
		FROM events e
		LEFT JOIN bookings b ON e.id = b.event_id AND b.status = 'confirmed'
		GROUP BY e.id, e.title, e.event_date
		ORDER BY revenue DESC
		LIMIT $1
	`
	mock.ExpectQuery(query).WithArgs(5).WillReturnRows(rows)

	result, err := repo.GetTopEvents(context.Background(), 5)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Event 1", result[0].Title)
}

func TestDashboardRepository_GetRecentActivities(t *testing.T) {
	repo, mock, cleanup := setupDashboardMock(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "action", "entity", "user_email", "created_at"}).
		AddRow(uuid.New(), "create", "event", "user@test.com", time.Now())

	query := `
		SELECT 
			al.id,
			al.action,
			al.entity_type as entity,
			u.email as user_email,
			al.created_at
		FROM audit_logs al
		LEFT JOIN users u ON al.user_id = u.id
		ORDER BY al.created_at DESC
		LIMIT $1
	`
	mock.ExpectQuery(query).WithArgs(5).WillReturnRows(rows)

	result, err := repo.GetRecentActivities(context.Background(), 5)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "user@test.com", result[0].UserEmail)
}
