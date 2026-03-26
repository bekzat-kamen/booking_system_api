package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
)

type PaymentRepository struct {
	db *sqlx.DB
}

func NewPaymentRepository(db *sqlx.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, payment *model.Payment) error {
	query := `
		INSERT INTO payment_transactions (id, booking_id, transaction_id, amount, status, payment_method, provider, provider_response, created_at, updated_at)
		VALUES (:id, :booking_id, :transaction_id, :amount, :status, :payment_method, :provider, :provider_response, :created_at, :updated_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":                payment.ID,
		"booking_id":        payment.BookingID,
		"transaction_id":    payment.TransactionID,
		"amount":            payment.Amount,
		"status":            payment.Status,
		"payment_method":    payment.PaymentMethod,
		"provider":          payment.Provider,
		"provider_response": payment.ProviderResponse,
		"created_at":        time.Now(),
		"updated_at":        time.Now(),
	})

	return err
}

func (r *PaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Payment, error) {
	query := `
		SELECT id, booking_id, transaction_id, amount, status, payment_method, provider, provider_response, paid_at, created_at, updated_at
		FROM payment_transactions
		WHERE id = $1
	`

	payment := &model.Payment{}
	err := r.db.GetContext(ctx, payment, query, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	return payment, nil
}

func (r *PaymentRepository) GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*model.Payment, error) {
	query := `
		SELECT id, booking_id, transaction_id, amount, status, payment_method, provider, provider_response, paid_at, created_at, updated_at
		FROM payment_transactions
		WHERE booking_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	payment := &model.Payment{}
	err := r.db.GetContext(ctx, payment, query, bookingID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	return payment, nil
}

func (r *PaymentRepository) GetByTransactionID(ctx context.Context, transactionID string) (*model.Payment, error) {
	query := `
		SELECT id, booking_id, transaction_id, amount, status, payment_method, provider, provider_response, paid_at, created_at, updated_at
		FROM payment_transactions
		WHERE transaction_id = $1
	`

	payment := &model.Payment{}
	err := r.db.GetContext(ctx, payment, query, transactionID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	return payment, nil
}

func (r *PaymentRepository) Update(ctx context.Context, payment *model.Payment) error {
	query := `
		UPDATE payment_transactions
		SET status = :status, transaction_id = :transaction_id, provider_response = :provider_response,
		    paid_at = :paid_at, updated_at = :updated_at
		WHERE id = :id
	`

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":                payment.ID,
		"status":            payment.Status,
		"transaction_id":    payment.TransactionID,
		"provider_response": payment.ProviderResponse,
		"paid_at":           payment.PaidAt,
		"updated_at":        time.Now(),
	})

	return err
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.PaymentStatus) error {
	query := `UPDATE payment_transactions SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	return err
}

func (r *PaymentRepository) SetProviderResponse(ctx context.Context, id uuid.UUID, response interface{}) error {
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return err
	}

	query := `UPDATE payment_transactions SET provider_response = $1, updated_at = $2 WHERE id = $3`
	_, err = r.db.ExecContext(ctx, query, string(responseJSON), time.Now(), id)
	return err
}
