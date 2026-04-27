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

type AdminUserRepositoryInterface interface {
	GetAllUsers(ctx context.Context, page, limit int, status, role string) ([]*model.User, int, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	UpdateUserRole(ctx context.Context, id uuid.UUID, role model.Role) error
	UpdateUserStatus(ctx context.Context, id uuid.UUID, status model.Status) error
	GetUserStats(ctx context.Context) (map[string]int64, error)
	GetUserBookingsCount(ctx context.Context, userID uuid.UUID) (int64, error)
	GetUserSpentAmount(ctx context.Context, userID uuid.UUID) (float64, error)
}

type AdminUserRepository struct {
	db *sqlx.DB
}

func NewAdminUserRepository(db *sqlx.DB) *AdminUserRepository {
	return &AdminUserRepository{db: db}
}

func (r *AdminUserRepository) GetAllUsers(ctx context.Context, page, limit int, status, role string) ([]*model.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	query := `
		SELECT id, email, password_hash, name, role, status, email_verified, created_at, updated_at
		FROM users
		WHERE ($1 = '' OR status = $1)
		  AND ($2 = '' OR role = $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	var users []*model.User
	err := r.db.SelectContext(ctx, &users, query, status, role, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	countQuery := `
		SELECT COUNT(*) FROM users
		WHERE ($1 = '' OR status = $1)
		  AND ($2 = '' OR role = $2)
	`
	var total int
	err = r.db.GetContext(ctx, &total, countQuery, status, role)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *AdminUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, name, role, status, email_verified, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &model.User{}
	err := r.db.GetContext(ctx, user, query, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *AdminUserRepository) UpdateUserRole(ctx context.Context, id uuid.UUID, role model.Role) error {
	query := `UPDATE users SET role = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, role, time.Now(), id)
	return err
}
func (r *AdminUserRepository) UpdateUserStatus(ctx context.Context, id uuid.UUID, status model.Status) error {
	query := `UPDATE users SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	return err
}

func (r *AdminUserRepository) GetUserStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)
	queries := map[string]string{
		"total":     `SELECT COUNT(*) FROM users`,
		"active":    `SELECT COUNT(*) FROM users WHERE status = 'active'`,
		"blocked":   `SELECT COUNT(*) FROM users WHERE status = 'blocked'`,
		"pending":   `SELECT COUNT(*) FROM users WHERE status = 'pending_verification'`,
		"user":      `SELECT COUNT(*) FROM users WHERE role = 'user'`,
		"organizer": `SELECT COUNT(*) FROM users WHERE role = 'organizer'`,
		"admin":     `SELECT COUNT(*) FROM users WHERE role = 'admin'`,
		"today":     `SELECT COUNT(*) FROM users WHERE created_at >= NOW() - INTERVAL '1 day'`,
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

func (r *AdminUserRepository) GetUserBookingsCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM bookings WHERE user_id = $1`
	var count int64
	err := r.db.GetContext(ctx, &count, query, userID)
	return count, err
}

func (r *AdminUserRepository) GetUserSpentAmount(ctx context.Context, userID uuid.UUID) (float64, error) {
	query := `SELECT COALESCE(SUM(final_amount), 0) FROM bookings WHERE user_id = $1 AND status = 'confirmed'`
	var amount float64
	err := r.db.GetContext(ctx, &amount, query, userID)
	return amount, err
}
