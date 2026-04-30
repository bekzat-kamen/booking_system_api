package service

import (
	"context"
	"testing"
	"time"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type adminPromocodeRepositoryMock struct {
	mock.Mock
}

func (m *adminPromocodeRepositoryMock) GetAllPromocodes(ctx context.Context, page, limit int, isActive string) ([]*model.Promocode, int, error) {
	args := m.Called(ctx, page, limit, isActive)
	resp, _ := args.Get(0).([]*model.Promocode)
	return resp, args.Int(1), args.Error(2)
}

func (m *adminPromocodeRepositoryMock) GetPromocodeByID(ctx context.Context, id uuid.UUID) (*model.Promocode, error) {
	args := m.Called(ctx, id)
	resp, _ := args.Get(0).(*model.Promocode)
	return resp, args.Error(1)
}

func (m *adminPromocodeRepositoryMock) Update(ctx context.Context, promocode *model.Promocode) error {
	args := m.Called(ctx, promocode)
	return args.Error(0)
}

func (m *adminPromocodeRepositoryMock) GetPromocodeUsageStats(ctx context.Context, promocodeID uuid.UUID) (map[string]interface{}, error) {
	args := m.Called(ctx, promocodeID)
	resp, _ := args.Get(0).(map[string]interface{})
	return resp, args.Error(1)
}

func (m *adminPromocodeRepositoryMock) GetPromocodesStats(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	resp, _ := args.Get(0).(map[string]int64)
	return resp, args.Error(1)
}

func (m *adminPromocodeRepositoryMock) BulkDeactivatePromocodes(ctx context.Context, ids []uuid.UUID) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *adminPromocodeRepositoryMock) DeletePromocode(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestAdminPromocodeServiceGetPromocodeDetailSuccess(t *testing.T) {
	ctx := context.Background()
	repo := new(adminPromocodeRepositoryMock)
	svc := NewAdminPromocodeService(repo)
	id := uuid.New()

	repo.On("GetPromocodeByID", ctx, id).Return(&model.Promocode{ID: id, Code: "SALE10"}, nil).Once()
	repo.On("GetPromocodeUsageStats", ctx, id).Return(map[string]interface{}{"used_count": int64(3)}, nil).Once()

	detail, err := svc.GetPromocodeDetail(ctx, id)

	require.NoError(t, err)
	assert.NotNil(t, detail["promocode"])
	assert.NotNil(t, detail["statistics"])
}

func TestAdminPromocodeServiceUpdatePromocodeInvalidDiscount(t *testing.T) {
	ctx := context.Background()
	repo := new(adminPromocodeRepositoryMock)
	svc := NewAdminPromocodeService(repo)
	id := uuid.New()

	repo.On("GetPromocodeByID", ctx, id).Return(&model.Promocode{ID: id, DiscountType: model.DiscountTypePercent}, nil).Once()

	resp, err := svc.UpdatePromocode(ctx, id, &model.UpdatePromocodeRequest{DiscountValue: 101})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrInvalidDiscountValue)
}

func TestAdminPromocodeServiceBulkDeactivateEmpty(t *testing.T) {
	svc := NewAdminPromocodeService(new(adminPromocodeRepositoryMock))

	err := svc.BulkDeactivate(context.Background(), nil)

	require.Error(t, err)
	assert.EqualError(t, err, "no promocode ids provided")
}

func TestAdminPromocodeServiceExportPromocodesToCSVSuccess(t *testing.T) {
	ctx := context.Background()
	repo := new(adminPromocodeRepositoryMock)
	svc := NewAdminPromocodeService(repo)
	validUntil := time.Now().Add(time.Hour)
	now := time.Now()

	repo.On("GetAllPromocodes", ctx, 1, 10000, "").Return([]*model.Promocode{
		{
			ID:            uuid.New(),
			Code:          "SALE10",
			DiscountType:  model.DiscountTypePercent,
			DiscountValue: 10,
			MinAmount:     100,
			MaxUses:       5,
			UsedCount:     1,
			IsActive:      true,
			ValidFrom:     now,
			ValidUntil:    &validUntil,
			CreatedAt:     now,
		},
	}, 1, nil).Once()

	rows, err := svc.ExportPromocodesToCSV(ctx)

	require.NoError(t, err)
	assert.Len(t, rows, 2)
	assert.Equal(t, "Code", rows[0][0])
	assert.Equal(t, "SALE10", rows[1][0])
}
