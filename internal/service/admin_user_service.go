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
	userRepo repository.AdminUserRepositoryInterface
}

func NewAdminUserService(userRepo repository.AdminUserRepositoryInterface) *AdminUserService {
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

func (s *AdminUserService) GetUserDetail(ctx context.Context, id uuid.UUID) (*model.UserDetail, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	bookingsCount, _ := s.userRepo.GetUserBookingsCount(ctx, id)
	spentAmount, _ := s.userRepo.GetUserSpentAmount(ctx, id)

	detail := &model.UserDetail{
		User: *user,
		Statistics: map[string]interface{}{
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

func (s *AdminUserService) GetUserStats(ctx context.Context) (*model.UserStats, error) {
	stats, err := s.userRepo.GetUserStats(ctx)
	if err != nil {
		return nil, err
	}

	return &model.UserStats{
		TotalUsers:     stats["total"],
		ActiveUsers:    stats["active"],
		BlockedUsers:   stats["blocked"],
		UnverifiedUsers: stats["pending"],
	}, nil
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
