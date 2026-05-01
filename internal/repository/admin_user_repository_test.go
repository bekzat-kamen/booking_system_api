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

func setupAdminUserMock(t *testing.T) (*AdminUserRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewAdminUserRepository(sqlxDB)

	return repo, mock, func() {
		_ = db.Close()
	}
}

func TestAdminUserRepository_GetAllUsers(t *testing.T) {
	repo, mock, cleanup := setupAdminUserMock(t)
	defer cleanup()

	t.Run("Success", func(t *testing.T) {
		query := `
		SELECT id, email, password_hash, name, role, status, email_verified, created_at, updated_at
		FROM users
		WHERE ($1 = '' OR status = $1)
		  AND ($2 = '' OR role = $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
		mock.ExpectQuery(query).
			WithArgs("active", "user", 10, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow(uuid.New(), "test@test.com"))

		countQuery := `
		SELECT COUNT(*) FROM users
		WHERE ($1 = '' OR status = $1)
		  AND ($2 = '' OR role = $2)
	`
		mock.ExpectQuery(countQuery).
			WithArgs("active", "user").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		users, total, err := repo.GetAllUsers(context.Background(), 1, 10, "active", "user")
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, users, 1)
	})
}

func TestAdminUserRepository_GetUserStats(t *testing.T) {
	repo, mock, cleanup := setupAdminUserMock(t)
	defer cleanup()

	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery("SELECT COUNT(*) FROM users").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	mock.ExpectQuery("SELECT COUNT(*) FROM users WHERE status = 'active'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery("SELECT COUNT(*) FROM users WHERE status = 'blocked'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery("SELECT COUNT(*) FROM users WHERE status = 'pending_verification'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery("SELECT COUNT(*) FROM users WHERE role = 'user'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(8))
	mock.ExpectQuery("SELECT COUNT(*) FROM users WHERE role = 'organizer'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT COUNT(*) FROM users WHERE role = 'admin'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT COUNT(*) FROM users WHERE created_at >= NOW() - INTERVAL '1 day'").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	stats, err := repo.GetUserStats(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(10), stats["total"])
}

func TestAdminUserRepository_GetUserSpentAmount(t *testing.T) {
	repo, mock, cleanup := setupAdminUserMock(t)
	defer cleanup()

	userID := uuid.New()

	mock.ExpectQuery("SELECT COALESCE(SUM(final_amount), 0) FROM bookings WHERE user_id = $1 AND status = 'confirmed'").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(150.50))

	amount, err := repo.GetUserSpentAmount(context.Background(), userID)
	assert.NoError(t, err)
	assert.Equal(t, 150.50, amount)
}
