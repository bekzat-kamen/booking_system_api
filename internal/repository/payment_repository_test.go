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

func setupPaymentMock(t *testing.T) (*PaymentRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewPaymentRepository(sqlxDB)

	return repo, mock, func() {
		db.Close()
	}
}

func TestPaymentRepository_Create(t *testing.T) {
	repo, mock, cleanup := setupPaymentMock(t)
	defer cleanup()

	payment := &model.Payment{
		ID:            uuid.New(),
		BookingID:     uuid.New(),
		TransactionID: "tx-123",
		Amount:        100.0,
		Status:        model.PaymentStatusPending,
		PaymentMethod: "card",
		Provider:      "stripe",
	}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO payment_transactions")).
		WithArgs(
			payment.ID,
			payment.BookingID,
			payment.TransactionID,
			payment.Amount,
			payment.Status,
			payment.PaymentMethod,
			payment.Provider,
			payment.ProviderResponse,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), payment)
	assert.NoError(t, err)
}

func TestPaymentRepository_GetByID(t *testing.T) {
	repo, mock, cleanup := setupPaymentMock(t)
	defer cleanup()

	paymentID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "booking_id", "transaction_id", "amount", "status", "payment_method", "provider", "provider_response", "paid_at", "created_at", "updated_at"}).
			AddRow(paymentID, uuid.New(), "tx-123", 100.0, "pending", "card", "stripe", "", nil, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, booking_id, transaction_id, amount, status, payment_method, provider, provider_response, paid_at, created_at, updated_at FROM payment_transactions WHERE id = $1")).
			WithArgs(paymentID).
			WillReturnRows(rows)

		payment, err := repo.GetByID(context.Background(), paymentID)
		require.NoError(t, err)
		assert.Equal(t, paymentID, payment.ID)
	})

	t.Run("Not Found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, booking_id, transaction_id, amount, status, payment_method, provider, provider_response, paid_at, created_at, updated_at FROM payment_transactions WHERE id = $1")).
			WithArgs(paymentID).
			WillReturnError(sql.ErrNoRows)

		payment, err := repo.GetByID(context.Background(), paymentID)
		assert.ErrorIs(t, err, ErrPaymentNotFound)
		assert.Nil(t, payment)
	})
}

func TestPaymentRepository_UpdateStatus(t *testing.T) {
	repo, mock, cleanup := setupPaymentMock(t)
	defer cleanup()

	paymentID := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE payment_transactions SET status = $1")).
		WithArgs(model.PaymentStatusSuccess, sqlmock.AnyArg(), paymentID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateStatus(context.Background(), paymentID, model.PaymentStatusSuccess)
	assert.NoError(t, err)
}
