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

func setupAdminPromocodeMock(t *testing.T) (*AdminPromocodeRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewAdminPromocodeRepository(sqlxDB)

	return repo, mock, func() {
		_ = db.Close()
	}
}

func TestAdminPromocodeRepository_GetAllPromocodes(t *testing.T) {
	repo, mock, cleanup := setupAdminPromocodeMock(t)
	defer cleanup()

	t.Run("Success", func(t *testing.T) {
		query := `
		SELECT id, code, description, discount_type, discount_value, min_amount,
		       max_uses, used_count, valid_from, valid_until, is_active,
		       created_by, created_at, updated_at
		FROM promocodes
		WHERE ($1 = '' OR is_active::text = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
		mock.ExpectQuery(query).
			WithArgs("true", 20, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).AddRow(uuid.New(), "SAVE10"))

		countQuery := `
		SELECT COUNT(*) FROM promocodes
		WHERE ($1 = '' OR is_active::text = $1)
	`
		mock.ExpectQuery(countQuery).
			WithArgs("true").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		promocodes, total, err := repo.GetAllPromocodes(context.Background(), 1, 20, "true")
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, promocodes, 1)
	})
}

func TestAdminPromocodeRepository_GetPromocodesStats(t *testing.T) {
	repo, mock, cleanup := setupAdminPromocodeMock(t)
	defer cleanup()

	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery("SELECT COUNT(*) FROM promocodes").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	mock.ExpectQuery("SELECT COUNT(*) FROM promocodes WHERE is_active = true").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(8))
	mock.ExpectQuery("SELECT COUNT(*) FROM promocodes WHERE is_active = false").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery("SELECT COUNT(*) FROM promocodes WHERE valid_until < NOW()").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT COALESCE(SUM(used_count), 0) FROM promocodes").WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(50))

	stats, err := repo.GetPromocodesStats(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(10), stats["total"])
	assert.Equal(t, int64(8), stats["active"])
}
