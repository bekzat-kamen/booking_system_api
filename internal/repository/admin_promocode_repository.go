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

type AdminPromocodeRepository struct {
	db *sqlx.DB
}

func NewAdminPromocodeRepository(db *sqlx.DB) *AdminPromocodeRepository {
	return &AdminPromocodeRepository{db: db}
}

func (r *AdminPromocodeRepository) GetAllPromocodes(ctx context.Context, page, limit int, isActive string) ([]*model.Promocode, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := `
		SELECT id, code, description, discount_type, discount_value, min_amount,
		       max_uses, used_count, valid_from, valid_until, is_active,
		       created_by, created_at, updated_at
		FROM promocodes
		WHERE ($1 = '' OR is_active::text = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var promocodes []*model.Promocode
	err := r.db.SelectContext(ctx, &promocodes, query, isActive, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	countQuery := `
		SELECT COUNT(*) FROM promocodes
		WHERE ($1 = '' OR is_active::text = $1)
	`
	var total int
	err = r.db.GetContext(ctx, &total, countQuery, isActive)
	if err != nil {
		return nil, 0, err
	}

	return promocodes, total, nil
}

func (r *AdminPromocodeRepository) GetPromocodeByID(ctx context.Context, id uuid.UUID) (*model.Promocode, error) {
	query := `
		SELECT id, code, description, discount_type, discount_value, min_amount,
		       max_uses, used_count, valid_from, valid_until, is_active,
		       created_by, created_at, updated_at
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

func (r *AdminPromocodeRepository) Update(ctx context.Context, promocode *model.Promocode) error {
	query := `
		UPDATE promocodes
		SET description = :description, discount_value = :discount_value, min_amount = :min_amount,
		    max_uses = :max_uses, valid_until = :valid_until, is_active = :is_active, updated_at = :updated_at
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":             promocode.ID,
		"description":    promocode.Description,
		"discount_value": promocode.DiscountValue,
		"min_amount":     promocode.MinAmount,
		"max_uses":       promocode.MaxUses,
		"valid_until":    promocode.ValidUntil,
		"is_active":      promocode.IsActive,
		"updated_at":     time.Now(),
	})
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrPromocodeNotFound
	}

	return nil
}

func (r *AdminPromocodeRepository) GetPromocodeUsageStats(ctx context.Context, promocodeID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var usedCount int64
	if err := r.db.GetContext(ctx, &usedCount, `SELECT used_count FROM promocodes WHERE id = $1`, promocodeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPromocodeNotFound
		}
		return nil, err
	}
	stats["used_count"] = usedCount

	var revenue float64
	if err := r.db.GetContext(ctx, &revenue, `SELECT COALESCE(SUM(final_amount), 0) FROM bookings WHERE status = 'confirmed'`); err != nil {
		return nil, err
	}
	stats["total_revenue"] = revenue

	var bookings int64
	if err := r.db.GetContext(ctx, &bookings, `SELECT COUNT(*) FROM bookings WHERE status = 'confirmed'`); err != nil {
		return nil, err
	}
	stats["total_bookings"] = bookings

	return stats, nil
}

func (r *AdminPromocodeRepository) GetPromocodesStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	queries := map[string]string{
		"total":      `SELECT COUNT(*) FROM promocodes`,
		"active":     `SELECT COUNT(*) FROM promocodes WHERE is_active = true`,
		"inactive":   `SELECT COUNT(*) FROM promocodes WHERE is_active = false`,
		"expired":    `SELECT COUNT(*) FROM promocodes WHERE valid_until < NOW()`,
		"total_uses": `SELECT COALESCE(SUM(used_count), 0) FROM promocodes`,
	}

	for key, query := range queries {
		var value int64
		if err := r.db.GetContext(ctx, &value, query); err != nil {
			return nil, err
		}
		stats[key] = value
	}

	return stats, nil
}

func (r *AdminPromocodeRepository) BulkDeactivatePromocodes(ctx context.Context, ids []uuid.UUID) error {
	query := `UPDATE promocodes SET is_active = false, updated_at = $1 WHERE id = ANY($2)`
	result, err := r.db.ExecContext(ctx, query, time.Now(), ids)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrPromocodeNotFound
	}

	return nil
}

func (r *AdminPromocodeRepository) DeletePromocode(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM promocodes WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrPromocodeNotFound
	}

	return nil
}
