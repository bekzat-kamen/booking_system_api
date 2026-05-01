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

func setupPromocodeMock(t *testing.T) (*PromocodeRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewPromocodeRepository(sqlxDB)

	return repo, mock, func() {
		_ = db.Close()
	}
}

func TestPromocodeRepository_Create(t *testing.T) {
	repo, mock, cleanup := setupPromocodeMock(t)
	defer cleanup()

	promocode := &model.Promocode{
		ID:            uuid.New(),
		Code:          "SAVE10",
		DiscountType:  model.DiscountTypePercent,
		DiscountValue: 10,
		IsActive:      true,
		ValidFrom:     time.Now(),
	}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO promocodes")).
		WithArgs(
			promocode.ID,
			promocode.Code,
			promocode.Description,
			promocode.DiscountType,
			promocode.DiscountValue,
			promocode.MinAmount,
			promocode.MaxUses,
			promocode.UsedCount,
			promocode.ValidFrom,
			promocode.ValidUntil,
			promocode.IsActive,
			promocode.CreatedBy,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), promocode)
	assert.NoError(t, err)
}

func TestPromocodeRepository_GetByCode(t *testing.T) {
	repo, mock, cleanup := setupPromocodeMock(t)
	defer cleanup()

	code := "SAVE10"
	promoID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "code", "description", "discount_type", "discount_value", "min_amount", "max_uses", "used_count", "valid_from", "valid_until", "is_active", "created_by", "created_at", "updated_at"}).
			AddRow(promoID, code, "Desc", "percent", 10.0, 0, 100, 0, time.Now(), nil, true, nil, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, code, description, discount_type, discount_value, min_amount, max_uses, used_count, valid_from, valid_until, is_active, created_by, created_at, updated_at FROM promocodes WHERE code = $1")).
			WithArgs(code).
			WillReturnRows(rows)

		promocode, err := repo.GetByCode(context.Background(), code)
		require.NoError(t, err)
		assert.Equal(t, promoID, promocode.ID)
		assert.Equal(t, code, promocode.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, code, description, discount_type, discount_value, min_amount, max_uses, used_count, valid_from, valid_until, is_active, created_by, created_at, updated_at FROM promocodes WHERE code = $1")).
			WithArgs(code).
			WillReturnError(sql.ErrNoRows)

		promocode, err := repo.GetByCode(context.Background(), code)
		assert.ErrorIs(t, err, ErrPromocodeNotFound)
		assert.Nil(t, promocode)
	})
}

func TestPromocodeRepository_IncrementUseCount(t *testing.T) {
	repo, mock, cleanup := setupPromocodeMock(t)
	defer cleanup()

	promoID := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE promocodes SET used_count = used_count + 1")).
		WithArgs(promoID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.IncrementUseCount(context.Background(), promoID)
	assert.NoError(t, err)
}

func TestPromocodeRepository_GetByID(t *testing.T) {
	repo, mock, cleanup := setupPromocodeMock(t)
	defer cleanup()

	promoID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, code, description, discount_type, discount_value, min_amount, max_uses, used_count, valid_from, valid_until, is_active, created_by, created_at, updated_at FROM promocodes WHERE id = $1")).
		WithArgs(promoID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(promoID))

	promo, err := repo.GetByID(context.Background(), promoID)
	assert.NoError(t, err)
	assert.Equal(t, promoID, promo.ID)
}

func TestPromocodeRepository_GetAll(t *testing.T) {
	repo, mock, cleanup := setupPromocodeMock(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, code, description, discount_type, discount_value, min_amount, max_uses, used_count, valid_from, valid_until, is_active, created_by, created_at, updated_at FROM promocodes ORDER BY created_at DESC LIMIT $1 OFFSET $2")).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

	promos, err := repo.GetAll(context.Background(), 10, 0)
	assert.NoError(t, err)
	assert.Len(t, promos, 1)
}

func TestPromocodeRepository_Update(t *testing.T) {
	repo, mock, cleanup := setupPromocodeMock(t)
	defer cleanup()

	promo := &model.Promocode{
		ID:   uuid.New(),
		Code: "NEWCODE",
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE promocodes SET description = $1, discount_value = $2, min_amount = $3, max_uses = $4, valid_until = $5, is_active = $6, updated_at = $7 WHERE id = $8")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), promo)
	assert.NoError(t, err)
}

func TestPromocodeRepository_Delete(t *testing.T) {
	repo, mock, cleanup := setupPromocodeMock(t)
	defer cleanup()

	promoID := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM promocodes WHERE id = $1")).
		WithArgs(promoID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), promoID)
	assert.NoError(t, err)
}

func TestPromocodeRepository_Count(t *testing.T) {
	repo, mock, cleanup := setupPromocodeMock(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM promocodes")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	count, err := repo.Count(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 10, count)
}

