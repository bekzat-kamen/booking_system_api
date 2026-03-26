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
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, name, role, status, email_verified, created_at, updated_at)
		VALUES (:id, :email, :password_hash, :name, :role, :status, :email_verified, :created_at, :updated_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":             user.ID,
		"email":          user.Email,
		"password_hash":  user.PasswordHash,
		"name":           user.Name,
		"role":           user.Role,
		"status":         user.Status,
		"email_verified": user.EmailVerified,
		"created_at":     time.Now(),
		"updated_at":     time.Now(),
	})

	if err != nil {
		if isUniqueViolation(err) {
			return ErrUserAlreadyExists
		}
		return err
	}

	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, name, role, status, email_verified, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &model.User{}
	err := r.db.GetContext(ctx, user, query, email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	if err := r.db.GetContext(ctx, &exists, query, email); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
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

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET email = :email, name = :name, role = :role, status = :status, 
		    email_verified = :email_verified, updated_at = :updated_at
		WHERE id = :id
	`

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":             user.ID,
		"email":          user.Email,
		"name":           user.Name,
		"role":           user.Role,
		"status":         user.Status,
		"email_verified": user.EmailVerified,
		"updated_at":     time.Now(),
	})

	return err
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, passwordHash, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	return err.Error() != "" &&
		(err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" ||
			err.Error() == "duplicate key value violates unique constraint")
}
