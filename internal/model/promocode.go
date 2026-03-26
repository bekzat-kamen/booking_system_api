package model

import (
	"time"

	"github.com/google/uuid"
)

type DiscountType string

const (
	DiscountTypePercent DiscountType = "percent"
	DiscountTypeFixed   DiscountType = "fixed"
)

type Promocode struct {
	ID            uuid.UUID    `db:"id" json:"id"`
	Code          string       `db:"code" json:"code"`
	Description   string       `db:"description" json:"description"`
	DiscountType  DiscountType `db:"discount_type" json:"discount_type"`
	DiscountValue float64      `db:"discount_value" json:"discount_value"`
	MinAmount     float64      `db:"min_amount" json:"min_amount"`
	MaxUses       int          `db:"max_uses" json:"max_uses"`
	UsedCount     int          `db:"used_count" json:"used_count"`
	ValidFrom     time.Time    `db:"valid_from" json:"valid_from"`
	ValidUntil    *time.Time   `db:"valid_until" json:"valid_until,omitempty"`
	IsActive      bool         `db:"is_active" json:"is_active"`
	CreatedBy     *uuid.UUID   `db:"created_by" json:"created_by,omitempty"`
	CreatedAt     time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time    `db:"updated_at" json:"updated_at"`
}

type CreatePromocodeRequest struct {
	Code          string       `json:"code" binding:"required,uppercase,alphanum,max=50"`
	Description   string       `json:"description" binding:"omitempty,max=500"`
	DiscountType  DiscountType `json:"discount_type" binding:"required,oneof=percent fixed"`
	DiscountValue float64      `json:"discount_value" binding:"required,min=0"`
	MinAmount     float64      `json:"min_amount" binding:"omitempty,min=0"`
	MaxUses       int          `json:"max_uses" binding:"omitempty,min=0"`
	ValidUntil    string       `json:"valid_until" binding:"omitempty"` // RFC3339 формат
}

type UpdatePromocodeRequest struct {
	Description   string  `json:"description" binding:"omitempty,max=500"`
	DiscountValue float64 `json:"discount_value" binding:"omitempty,min=0"`
	MinAmount     float64 `json:"min_amount" binding:"omitempty,min=0"`
	MaxUses       int     `json:"max_uses" binding:"omitempty,min=0"`
	ValidUntil    string  `json:"valid_until" binding:"omitempty"`
	IsActive      *bool   `json:"is_active" binding:"omitempty"`
}

type ValidatePromocodeRequest struct {
	Code        string    `json:"code" binding:"required"`
	EventID     uuid.UUID `json:"event_id" binding:"required"`
	TotalAmount float64   `json:"total_amount" binding:"required,min=0"`
}

type PromocodeResponse struct {
	ID            uuid.UUID    `json:"id"`
	Code          string       `json:"code"`
	Description   string       `json:"description"`
	DiscountType  DiscountType `json:"discount_type"`
	DiscountValue float64      `json:"discount_value"`
	MinAmount     float64      `json:"min_amount"`
	MaxUses       int          `json:"max_uses"`
	UsedCount     int          `json:"used_count"`
	ValidFrom     time.Time    `json:"valid_from"`
	ValidUntil    *time.Time   `json:"valid_until,omitempty"`
	IsActive      bool         `json:"is_active"`
}

type PromocodeValidationResponse struct {
	Valid          bool    `json:"valid"`
	Code           string  `json:"code"`
	DiscountType   string  `json:"discount_type"`
	DiscountValue  float64 `json:"discount_value"`
	OriginalAmount float64 `json:"original_amount"`
	DiscountAmount float64 `json:"discount_amount"`
	FinalAmount    float64 `json:"final_amount"`
	Message        string  `json:"message,omitempty"`
}
