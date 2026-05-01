package service

import (
	"context"
	"testing"
	"time"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type promocodeRepositoryMock struct {
	mock.Mock
}

func (m *promocodeRepositoryMock) Create(ctx context.Context, promocode *model.Promocode) error {
	args := m.Called(ctx, promocode)
	return args.Error(0)
}

func (m *promocodeRepositoryMock) GetByCode(ctx context.Context, code string) (*model.Promocode, error) {
	args := m.Called(ctx, code)
	resp, _ := args.Get(0).(*model.Promocode)
	return resp, args.Error(1)
}

func (m *promocodeRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*model.Promocode, error) {
	args := m.Called(ctx, id)
	resp, _ := args.Get(0).(*model.Promocode)
	return resp, args.Error(1)
}

func (m *promocodeRepositoryMock) GetAll(ctx context.Context, limit, offset int) ([]*model.Promocode, error) {
	args := m.Called(ctx, limit, offset)
	resp, _ := args.Get(0).([]*model.Promocode)
	return resp, args.Error(1)
}

func (m *promocodeRepositoryMock) Update(ctx context.Context, promocode *model.Promocode) error {
	args := m.Called(ctx, promocode)
	return args.Error(0)
}

func (m *promocodeRepositoryMock) IncrementUseCount(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *promocodeRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *promocodeRepositoryMock) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func TestPromocodeServiceCreatePromocodeInvalidPercent(t *testing.T) {
	svc := NewPromocodeService(new(promocodeRepositoryMock))

	promocode, err := svc.CreatePromocode(context.Background(), uuid.New(), &model.CreatePromocodeRequest{
		Code:          "SALE",
		DiscountType:  model.DiscountTypePercent,
		DiscountValue: 120,
	})

	require.Error(t, err)
	assert.Nil(t, promocode)
	assert.ErrorIs(t, err, ErrInvalidDiscountValue)
}

func TestPromocodeServiceCreatePromocodeSuccess(t *testing.T) {
	ctx := context.Background()
	repo := new(promocodeRepositoryMock)
	svc := NewPromocodeService(repo)
	userID := uuid.New()

	repo.On("Create", ctx, mock.MatchedBy(func(p *model.Promocode) bool {
		return p.Code == "SALE10" && p.CreatedBy != nil && *p.CreatedBy == userID && p.IsActive
	})).Return(nil).Once()

	promocode, err := svc.CreatePromocode(ctx, userID, &model.CreatePromocodeRequest{
		Code:          "SALE10",
		DiscountType:  model.DiscountTypePercent,
		DiscountValue: 10,
	})

	require.NoError(t, err)
	require.NotNil(t, promocode)
	assert.Equal(t, "SALE10", promocode.Code)
	repo.AssertExpectations(t)
}

func TestPromocodeServiceValidatePromocodeNotFound(t *testing.T) {
	ctx := context.Background()
	repo := new(promocodeRepositoryMock)
	svc := NewPromocodeService(repo)

	repo.On("GetByCode", ctx, "MISSING").Return((*model.Promocode)(nil), repository.ErrPromocodeNotFound).Once()

	resp, err := svc.ValidatePromocode(ctx, &model.ValidatePromocodeRequest{Code: "MISSING", TotalAmount: 100})

	require.NoError(t, err)
	assert.False(t, resp.Valid)
	assert.Equal(t, "promocode not found", resp.Message)
}

func TestPromocodeServiceValidatePromocodeSuccess(t *testing.T) {
	ctx := context.Background()
	repo := new(promocodeRepositoryMock)
	svc := NewPromocodeService(repo)
	now := time.Now()

	repo.On("GetByCode", ctx, "SALE10").Return(&model.Promocode{
		ID:            uuid.New(),
		Code:          "SALE10",
		DiscountType:  model.DiscountTypePercent,
		DiscountValue: 10,
		MinAmount:     50,
		ValidFrom:     now.Add(-time.Hour),
		IsActive:      true,
	}, nil).Once()

	resp, err := svc.ValidatePromocode(ctx, &model.ValidatePromocodeRequest{Code: "SALE10", TotalAmount: 200})

	require.NoError(t, err)
	assert.True(t, resp.Valid)
	assert.Equal(t, 20.0, resp.DiscountAmount)
	assert.Equal(t, 180.0, resp.FinalAmount)
}

func TestPromocodeServiceUpdatePromocodeInvalidDiscount(t *testing.T) {
	ctx := context.Background()
	repo := new(promocodeRepositoryMock)
	svc := NewPromocodeService(repo)
	id := uuid.New()

	repo.On("GetByID", ctx, id).Return(&model.Promocode{
		ID:           id,
		DiscountType: model.DiscountTypePercent,
	}, nil).Once()

	resp, err := svc.UpdatePromocode(ctx, id, &model.UpdatePromocodeRequest{DiscountValue: 150})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrInvalidDiscountValue)
}

func TestPromocodeServiceDeactivatePromocodeSuccess(t *testing.T) {
	ctx := context.Background()
	repo := new(promocodeRepositoryMock)
	svc := NewPromocodeService(repo)
	id := uuid.New()
	promocode := &model.Promocode{ID: id, IsActive: true}

	repo.On("GetByID", ctx, id).Return(promocode, nil).Once()
	repo.On("Update", ctx, promocode).Return(nil).Once()

	err := svc.DeactivatePromocode(ctx, id)

	require.NoError(t, err)
	assert.False(t, promocode.IsActive)
	repo.AssertExpectations(t)
}

func TestPromocodeServiceGetPromocode(t *testing.T) {
	ctx := context.Background()
	repo := new(promocodeRepositoryMock)
	svc := NewPromocodeService(repo)
	id := uuid.New()

	repo.On("GetByID", ctx, id).Return(&model.Promocode{ID: id, Code: "TEST"}, nil).Once()

	resp, err := svc.GetPromocode(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, id, resp.ID)
}

func TestPromocodeServiceGetAllPromocodes(t *testing.T) {
	ctx := context.Background()
	repo := new(promocodeRepositoryMock)
	svc := NewPromocodeService(repo)

	repo.On("GetAll", ctx, 10, 0).Return([]*model.Promocode{{ID: uuid.New()}}, nil).Once()
	repo.On("Count", ctx).Return(1, nil).Once()

	resp, total, err := svc.GetAllPromocodes(ctx, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, resp, 1)
}
func TestPromocodeServiceApplyPromocodeSuccess(t *testing.T) {
	ctx := context.Background()
	repo := new(promocodeRepositoryMock)
	svc := NewPromocodeService(repo)
	code := "SALE10"
	id := uuid.New()

	repo.On("GetByCode", ctx, code).Return(&model.Promocode{ID: id, Code: code}, nil).Once()
	repo.On("IncrementUseCount", ctx, id).Return(nil).Once()

	err := svc.ApplyPromocode(ctx, code)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestPromocodeServiceDeletePromocodeSuccess(t *testing.T) {
	ctx := context.Background()
	repo := new(promocodeRepositoryMock)
	svc := NewPromocodeService(repo)
	id := uuid.New()

	repo.On("Delete", ctx, id).Return(nil).Once()

	err := svc.DeletePromocode(ctx, id)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}
