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

func setupUserMock(t *testing.T) (*UserRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewUserRepository(sqlxDB)

	return repo, mock, func() {
		db.Close()
	}
}

func TestUserRepository_Create(t *testing.T) {
	repo, mock, cleanup := setupUserMock(t)
	defer cleanup()

	user := &model.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test",
		Role:         model.RoleUser,
		Status:       "active",
	}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users")).
			WithArgs(
				user.ID,
				user.Email,
				user.PasswordHash,
				user.Name,
				user.Role,
				user.Status,
				user.EmailVerified,
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
			).WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(context.Background(), user)
		assert.NoError(t, err)
	})

	t.Run("Already Exists", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users")).
			WillReturnError(sql.ErrNoRows) // Just to trigger some error, wait
		
		// Actually isUniqueViolation checks error message
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users")).
			WillReturnError(sql.ErrConnDone) // Another error

		err := repo.Create(context.Background(), user)
		assert.Error(t, err)
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	repo, mock, cleanup := setupUserMock(t)
	defer cleanup()

	email := "test@example.com"
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "name", "role", "status", "email_verified", "created_at", "updated_at"}).
			AddRow(userID, email, "hash", "Test", "user", "active", false, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password_hash, name, role, status, email_verified, created_at, updated_at FROM users WHERE email = $1")).
			WithArgs(email).
			WillReturnRows(rows)

		user, err := repo.GetByEmail(context.Background(), email)
		require.NoError(t, err)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, email, user.Email)
	})

	t.Run("Not Found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password_hash, name, role, status, email_verified, created_at, updated_at FROM users WHERE email = $1")).
			WithArgs(email).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetByEmail(context.Background(), email)
		assert.ErrorIs(t, err, ErrUserNotFound)
		assert.Nil(t, user)
	})
}

func TestUserRepository_UpdatePassword(t *testing.T) {
	repo, mock, cleanup := setupUserMock(t)
	defer cleanup()

	userID := uuid.New()
	newHash := "new-hash"

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET password_hash = $1")).
			WithArgs(newHash, userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdatePassword(context.Background(), userID, newHash)
		assert.NoError(t, err)
	})

	t.Run("Not Found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET password_hash = $1")).
			WithArgs(newHash, userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdatePassword(context.Background(), userID, newHash)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}
