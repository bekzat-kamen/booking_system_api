package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrPromocodeNotFound  = errors.New("promocode not found")
	ErrPromocodeExpired   = errors.New("promocode has expired")
	ErrPromocodeNotActive = errors.New("promocode is not active")
	ErrPromocodeMaxUses   = errors.New("promocode usage limit reached")
	ErrPromocodeMinAmount = errors.New("minimum amount not met")
)

type PromocodeRepository struct {
	db *sqlx.DB
}

func NewPromocodeRepository(db *sqlx.DB) *PromocodeRepository {
	return &PromocodeRepository{db: db}
}

func (r *PromocodeRepository) Create(ctx context.Context, promocode *model.Promocode) error {
	query := `
		INSERT INTO promocodes (id, code, description, discount_type, discount_value, min_amount, max_uses, used_count, valid_from, valid_until, is_active, created_by, created_at, updated_at)
		VALUES (:id, :code, :description, :discount_type, :discount_value, :min_amount, :max_uses, :used_count, :valid_from, :valid_until, :is_active, :created_by, :created_at, :updated_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":             promocode.ID,
		"code":           promocode.Code,
		"description":    promocode.Description,
		"discount_type":  promocode.DiscountType,
		"discount_value": promocode.DiscountValue,
		"min_amount":     promocode.MinAmount,
		"max_uses":       promocode.MaxUses,
		"used_count":     promocode.UsedCount,
		"valid_from":     promocode.ValidFrom,
		"valid_until":    promocode.ValidUntil,
		"is_active":      promocode.IsActive,
		"created_by":     promocode.CreatedBy,
		"created_at":     time.Now(),
		"updated_at":     time.Now(),
	})

	return err
}

func (r *PromocodeRepository) GetByCode(ctx context.Context, code string) (*model.Promocode, error) {
	query := `
		SELECT id, code, description, discount_type, discount_value, min_amount, max_uses, used_count, valid_from, valid_until, is_active, created_by, created_at, updated_at
		FROM promocodes
		WHERE code = $1
	`

	promocode := &model.Promocode{}
	err := r.db.GetContext(ctx, promocode, query, code)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPromocodeNotFound
		}
		return nil, err
	}

	return promocode, nil
}

func (r *PromocodeRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Promocode, error) {
	query := `
		SELECT id, code, description, discount_type, discount_value, min_amount, max_uses, used_count, valid_from, valid_until, is_active, created_by, created_at, updated_at
		FROM promocodes
		WHERE id = $1
	`

	promocode := &model.Promocode{}
	err := r.db.GetContext(ctx, promocode, query, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPromocodeNotFound
		}
		return nil, err
	}

	return promocode, nil
}

func (r *PromocodeRepository) GetAll(ctx context.Context, limit, offset int) ([]*model.Promocode, error) {
	query := `
		SELECT id, code, description, discount_type, discount_value, min_amount, max_uses, used_count, valid_from, valid_until, is_active, created_by, created_at, updated_at
		FROM promocodes
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var promocodes []*model.Promocode
	err := r.db.SelectContext(ctx, &promocodes, query, limit, offset)

	return promocodes, err
}

func (r *PromocodeRepository) Update(ctx context.Context, promocode *model.Promocode) error {
	query := `
		UPDATE promocodes
		SET description = :description, discount_value = :discount_value, min_amount = :min_amount,
		    max_uses = :max_uses, valid_until = :valid_until, is_active = :is_active, updated_at = :updated_at
		WHERE id = :id
	`

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":             promocode.ID,
		"description":    promocode.Description,
		"discount_value": promocode.DiscountValue,
		"min_amount":     promocode.MinAmount,
		"max_uses":       promocode.MaxUses,
		"valid_until":    promocode.ValidUntil,
		"is_active":      promocode.IsActive,
		"updated_at":     time.Now(),
	})

	return err
}

func (r *PromocodeRepository) IncrementUseCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE promocodes SET used_count = used_count + 1, updated_at = $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, time.Now())
	return err
}

func (r *PromocodeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM promocodes WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PromocodeRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM promocodes`
	var count int
	err := r.db.GetContext(ctx, &count, query)
	return count, err
}
