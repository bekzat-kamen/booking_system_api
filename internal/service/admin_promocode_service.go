package service

import (
	"context"
	"errors"
	"time"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/google/uuid"
)

type AdminPromocodeService struct {
	promocodeRepo *repository.AdminPromocodeRepository
}

func NewAdminPromocodeService(promocodeRepo *repository.AdminPromocodeRepository) *AdminPromocodeService {
	return &AdminPromocodeService{promocodeRepo: promocodeRepo}
}

func (s *AdminPromocodeService) GetAllPromocodes(ctx context.Context, page, limit int, isActive string) ([]*model.Promocode, int, error) {
	return s.promocodeRepo.GetAllPromocodes(ctx, page, limit, isActive)
}

func (s *AdminPromocodeService) GetPromocodeDetail(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) {
	promocode, err := s.promocodeRepo.GetPromocodeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	stats, _ := s.promocodeRepo.GetPromocodeUsageStats(ctx, id)

	detail := map[string]interface{}{
		"promocode": map[string]interface{}{
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
			"created_at":     promocode.CreatedAt,
			"updated_at":     promocode.UpdatedAt,
		},
		"statistics": stats,
	}

	return detail, nil
}

func (s *AdminPromocodeService) UpdatePromocode(ctx context.Context, id uuid.UUID, req *model.UpdatePromocodeRequest) (*model.Promocode, error) {
	promocode, err := s.promocodeRepo.GetPromocodeByID(ctx, id)
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

func (s *AdminPromocodeService) DeletePromocode(ctx context.Context, id uuid.UUID) error {
	return s.promocodeRepo.DeletePromocode(ctx, id)
}

func (s *AdminPromocodeService) BulkDeactivate(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return errors.New("no promocode ids provided")
	}
	return s.promocodeRepo.BulkDeactivatePromocodes(ctx, ids)
}

func (s *AdminPromocodeService) GetPromocodesStats(ctx context.Context) (map[string]int64, error) {
	return s.promocodeRepo.GetPromocodesStats(ctx)
}

func (s *AdminPromocodeService) ExportPromocodesToCSV(ctx context.Context) ([][]string, error) {
	promocodes, _, err := s.promocodeRepo.GetAllPromocodes(ctx, 1, 10000, "")
	if err != nil {
		return nil, err
	}

	var rows [][]string

	rows = append(rows, []string{
		"Code", "Description", "Discount Type", "Discount Value",
		"Min Amount", "Max Uses", "Used Count", "Is Active",
		"Valid From", "Valid Until", "Created At",
	})

	for _, p := range promocodes {
		validUntil := ""
		if p.ValidUntil != nil {
			validUntil = p.ValidUntil.Format(time.RFC3339)
		}

		rows = append(rows, []string{
			p.Code,
			p.Description,
			string(p.DiscountType),
			formatFloat(p.DiscountValue),
			formatFloat(p.MinAmount),
			intToString(p.MaxUses),
			intToString(p.UsedCount),
			boolToString(p.IsActive),
			p.ValidFrom.Format(time.RFC3339),
			validUntil,
			p.CreatedAt.Format(time.RFC3339),
		})
	}

	return rows, nil
}

func formatFloat(f float64) string {
	return string(rune(f))
}

func intToString(i int) string {
	return string(rune(i))
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
