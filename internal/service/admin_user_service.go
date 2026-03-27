package service

import (
	"context"
	"errors"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrCannotModifyAdmin   = errors.New("cannot modify admin user")
	ErrCannotBlockYourself = errors.New("cannot block yourself")
)

type AdminUserService struct {
	userRepo *repository.AdminUserRepository
}

func NewAdminUserService(userRepo *repository.AdminUserRepository) *AdminUserService {
	return &AdminUserService{userRepo: userRepo}
}

func (s *AdminUserService) GetAllUsers(ctx context.Context, page, limit int, status, role string) ([]*model.User, int, error) {
	return s.userRepo.GetAllUsers(ctx, page, limit, status, role)
}

func (s *AdminUserService) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = ""

	return user, nil
}

func (s *AdminUserService) GetUserDetail(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	bookingsCount, _ := s.userRepo.GetUserBookingsCount(ctx, id)
	spentAmount, _ := s.userRepo.GetUserSpentAmount(ctx, id)

	detail := map[string]interface{}{
		"user": map[string]interface{}{
			"id":             user.ID,
			"email":          user.Email,
			"name":           user.Name,
			"role":           user.Role,
			"status":         user.Status,
			"email_verified": user.EmailVerified,
			"created_at":     user.CreatedAt,
			"updated_at":     user.UpdatedAt,
		},
		"statistics": map[string]interface{}{
			"total_bookings": bookingsCount,
			"total_spent":    spentAmount,
		},
	}

	return detail, nil
}

func (s *AdminUserService) UpdateUserRole(ctx context.Context, id uuid.UUID, role model.Role, adminID uuid.UUID) error {

	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	if user.Role == model.RoleAdmin && id != adminID {
		return ErrCannotModifyAdmin
	}

	return s.userRepo.UpdateUserRole(ctx, id, role)
}

func (s *AdminUserService) BlockUser(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error {

	if id == adminID {
		return ErrCannotBlockYourself
	}

	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	if user.Role == model.RoleAdmin {
		return ErrCannotModifyAdmin
	}

	return s.userRepo.UpdateUserStatus(ctx, id, model.StatusBlocked)
}

func (s *AdminUserService) UnblockUser(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.UpdateUserStatus(ctx, id, model.StatusActive)
}

func (s *AdminUserService) GetUserStats(ctx context.Context) (map[string]int64, error) {
	return s.userRepo.GetUserStats(ctx)
}

func (s *AdminUserService) DeleteUser(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error {
	if id == adminID {
		return ErrCannotBlockYourself
	}

	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	if user.Role == model.RoleAdmin {
		return ErrCannotModifyAdmin
	}

	return s.userRepo.UpdateUserStatus(ctx, id, model.StatusBlocked)
}
