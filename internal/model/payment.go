package model

import (
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusSuccess  PaymentStatus = "success"
	PaymentStatusFailed   PaymentStatus = "failed"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

type PaymentMethod string

const (
	PaymentMethodCard PaymentMethod = "card"
	PaymentMethodSBP  PaymentMethod = "sbp"
	PaymentMethodCash PaymentMethod = "cash"
)

type Payment struct {
	ID               uuid.UUID     `db:"id" json:"id"`
	BookingID        uuid.UUID     `db:"booking_id" json:"booking_id"`
	TransactionID    string        `db:"transaction_id" json:"transaction_id,omitempty"`
	Amount           float64       `db:"amount" json:"amount"`
	Status           PaymentStatus `db:"status" json:"status"`
	PaymentMethod    PaymentMethod `db:"payment_method" json:"payment_method"`
	Provider         string        `db:"provider" json:"provider"`
	ProviderResponse string        `db:"provider_response" json:"-"`
	PaidAt           *time.Time    `db:"paid_at" json:"paid_at,omitempty"`
	CreatedAt        time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time     `db:"updated_at" json:"updated_at"`
}

type CreatePaymentRequest struct {
	BookingID     uuid.UUID     `json:"booking_id" binding:"required"`
	PaymentMethod PaymentMethod `json:"payment_method" binding:"required,oneof=card sbp cash"`
}

type PaymentResponse struct {
	ID            uuid.UUID     `json:"id"`
	BookingID     uuid.UUID     `json:"booking_id"`
	Amount        float64       `json:"amount"`
	Status        PaymentStatus `json:"status"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	PaidAt        *time.Time    `json:"paid_at,omitempty"`
	PaymentURL    string        `json:"payment_url,omitempty"`
}

type WebhookRequest struct {
	TransactionID string  `json:"transaction_id"`
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
	Signature     string  `json:"signature"`
}
