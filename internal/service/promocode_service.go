package service

import (
	"context"
	"errors"
	"time"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrInvalidDiscountValue = errors.New("invalid discount value")
)

type PromocodeService struct {
	promocodeRepo *repository.PromocodeRepository
}

func NewPromocodeService(promocodeRepo *repository.PromocodeRepository) *PromocodeService {
	return &PromocodeService{promocodeRepo: promocodeRepo}
}

func (s *PromocodeService) CreatePromocode(ctx context.Context, createdBy uuid.UUID, req *model.CreatePromocodeRequest) (*model.Promocode, error) {

	if req.DiscountType == model.DiscountTypePercent && (req.DiscountValue < 0 || req.DiscountValue > 100) {
		return nil, ErrInvalidDiscountValue
	}

	var validUntil *time.Time
	if req.ValidUntil != "" {
		parsed, err := time.Parse(time.RFC3339, req.ValidUntil)
		if err != nil {
			return nil, errors.New("invalid valid_until format, use RFC3339")
		}
		validUntil = &parsed
	}

	promocode := &model.Promocode{
		ID:            uuid.New(),
		Code:          req.Code,
		Description:   req.Description,
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		MinAmount:     req.MinAmount,
		MaxUses:       req.MaxUses,
		UsedCount:     0,
		ValidFrom:     time.Now(),
		ValidUntil:    validUntil,
		IsActive:      true,
		CreatedBy:     &createdBy,
	}
	if err := s.promocodeRepo.Create(ctx, promocode); err != nil {
		return nil, errors.New("failed to create promocode")
	}

	return promocode, nil
}

func (s *PromocodeService) ValidatePromocode(ctx context.Context, req *model.ValidatePromocodeRequest) (*model.PromocodeValidationResponse, error) {

	promocode, err := s.promocodeRepo.GetByCode(ctx, req.Code)
	if err != nil {
		return &model.PromocodeValidationResponse{
			Valid:   false,
			Message: "promocode not found",
		}, nil
	}

	if !promocode.IsActive {
		return &model.PromocodeValidationResponse{
			Valid:   false,
			Message: "promocode is not active",
		}, nil
	}

	now := time.Now()
	if now.Before(promocode.ValidFrom) {
		return &model.PromocodeValidationResponse{
			Valid:   false,
			Message: "promocode is not yet valid",
		}, nil
	}
	if promocode.ValidUntil != nil && now.After(*promocode.ValidUntil) {
		return &model.PromocodeValidationResponse{
			Valid:   false,
			Message: "promocode has expired",
		}, nil
	}

	if promocode.MaxUses > 0 && promocode.UsedCount >= promocode.MaxUses {
		return &model.PromocodeValidationResponse{
			Valid:   false,
			Message: "promocode usage limit reached",
		}, nil
	}

	if req.TotalAmount < promocode.MinAmount {
		return &model.PromocodeValidationResponse{
			Valid:          false,
			Message:        "minimum amount not met",
			OriginalAmount: req.TotalAmount,
		}, nil
	}

	discountAmount := 0.0
	if promocode.DiscountType == model.DiscountTypePercent {
		discountAmount = req.TotalAmount * (promocode.DiscountValue / 100)
	} else {
		discountAmount = promocode.DiscountValue
	}

	if discountAmount > req.TotalAmount {
		discountAmount = req.TotalAmount
	}

	finalAmount := req.TotalAmount - discountAmount

	return &model.PromocodeValidationResponse{
		Valid:          true,
		Code:           promocode.Code,
		DiscountType:   string(promocode.DiscountType),
		DiscountValue:  promocode.DiscountValue,
		OriginalAmount: req.TotalAmount,
		DiscountAmount: discountAmount,
		FinalAmount:    finalAmount,
	}, nil
}

func (s *PromocodeService) ApplyPromocode(ctx context.Context, code string) error {
	promocode, err := s.promocodeRepo.GetByCode(ctx, code)
	if err != nil {
		return err
	}

	return s.promocodeRepo.IncrementUseCount(ctx, promocode.ID)
}

func (s *PromocodeService) GetPromocode(ctx context.Context, id uuid.UUID) (*model.Promocode, error) {
	return s.promocodeRepo.GetByID(ctx, id)
}

func (s *PromocodeService) GetAllPromocodes(ctx context.Context, page, limit int) ([]*model.Promocode, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	promocodes, err := s.promocodeRepo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.promocodeRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return promocodes, total, nil
}

func (s *PromocodeService) UpdatePromocode(ctx context.Context, id uuid.UUID, req *model.UpdatePromocodeRequest) (*model.Promocode, error) {
	promocode, err := s.promocodeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Description != "" {
		promocode.Description = req.Description
	}
	if req.DiscountValue > 0 {
		if promocode.DiscountType == model.DiscountTypePercent && req.DiscountValue > 100 {
			return nil, ErrInvalidDiscountValue
		}
		promocode.DiscountValue = req.DiscountValue
	}
	if req.MinAmount > 0 {
		promocode.MinAmount = req.MinAmount
	}
	if req.MaxUses > 0 {
		promocode.MaxUses = req.MaxUses
	}
	if req.ValidUntil != "" {
		parsed, err := time.Parse(time.RFC3339, req.ValidUntil)
		if err != nil {
			return nil, errors.New("invalid valid_until format")
		}
		promocode.ValidUntil = &parsed
	}
	if req.IsActive != nil {
		promocode.IsActive = *req.IsActive
	}

	if err := s.promocodeRepo.Update(ctx, promocode); err != nil {
		return nil, errors.New("failed to update promocode")
	}

	return promocode, nil
}

func (s *PromocodeService) DeletePromocode(ctx context.Context, id uuid.UUID) error {
	return s.promocodeRepo.Delete(ctx, id)
}

func (s *PromocodeService) DeactivatePromocode(ctx context.Context, id uuid.UUID) error {
	promocode, err := s.promocodeRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	promocode.IsActive = false
	return s.promocodeRepo.Update(ctx, promocode)
}
