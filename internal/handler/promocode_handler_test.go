package handler

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type promocodeServiceMock struct {
	mock.Mock
}

func (m *promocodeServiceMock) CreatePromocode(ctx context.Context, createdBy uuid.UUID, req *model.CreatePromocodeRequest) (*model.Promocode, error) {
	args := m.Called(ctx, createdBy, req)
	resp, _ := args.Get(0).(*model.Promocode)
	return resp, args.Error(1)
}

func (m *promocodeServiceMock) ValidatePromocode(ctx context.Context, req *model.ValidatePromocodeRequest) (*model.PromocodeValidationResponse, error) {
	args := m.Called(ctx, req)
	resp, _ := args.Get(0).(*model.PromocodeValidationResponse)
	return resp, args.Error(1)
}

func (m *promocodeServiceMock) ApplyPromocode(ctx context.Context, code string) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

func (m *promocodeServiceMock) GetPromocode(ctx context.Context, id uuid.UUID) (*model.Promocode, error) {
	args := m.Called(ctx, id)
	resp, _ := args.Get(0).(*model.Promocode)
	return resp, args.Error(1)
}

func (m *promocodeServiceMock) GetAllPromocodes(ctx context.Context, page, limit int) ([]*model.Promocode, int, error) {
	args := m.Called(ctx, page, limit)
	resp, _ := args.Get(0).([]*model.Promocode)
	return resp, args.Int(1), args.Error(2)
}

func (m *promocodeServiceMock) UpdatePromocode(ctx context.Context, id uuid.UUID, req *model.UpdatePromocodeRequest) (*model.Promocode, error) {
	args := m.Called(ctx, id, req)
	resp, _ := args.Get(0).(*model.Promocode)
	return resp, args.Error(1)
}

func (m *promocodeServiceMock) DeletePromocode(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *promocodeServiceMock) DeactivatePromocode(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestPromocodeHandlerCreatePromocodeSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	promocodeSvc := new(promocodeServiceMock)
	h := NewPromocodeHandler(promocodeSvc)
	router := gin.New()
	router.Use(addUserContext())
	router.POST("/promocodes", h.CreatePromocode)

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	validUntil := time.Now().Add(24 * time.Hour)
	reqBody := map[string]interface{}{
		"code":           "SALE10",
		"discount_type":  "percent",
		"discount_value": 10,
		"min_amount":     100,
		"max_uses":       50,
		"valid_until":    validUntil.Format(time.RFC3339),
	}

	promocodeSvc.On("CreatePromocode", mock.Anything, userID, mock.MatchedBy(func(req *model.CreatePromocodeRequest) bool {
		return req.Code == "SALE10" && req.DiscountType == model.DiscountTypePercent && req.DiscountValue == 10
	})).Return(&model.Promocode{
		ID:            uuid.New(),
		Code:          "SALE10",
		DiscountType:  model.DiscountTypePercent,
		DiscountValue: 10,
		MinAmount:     100,
		MaxUses:       50,
		ValidUntil:    &validUntil,
		IsActive:      true,
	}, nil).Once()

	w := performJSONRequest(router, http.MethodPost, "/promocodes", reqBody)

	require.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "SALE10")
	promocodeSvc.AssertExpectations(t)
}

func TestPromocodeHandlerValidatePromocodeSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	promocodeSvc := new(promocodeServiceMock)
	h := NewPromocodeHandler(promocodeSvc)
	router := gin.New()
	router.POST("/promocodes/validate", h.ValidatePromocode)

	eventID := uuid.New()
	promocodeSvc.On("ValidatePromocode", mock.Anything, mock.MatchedBy(func(req *model.ValidatePromocodeRequest) bool {
		return req.Code == "SALE10" && req.EventID == eventID && req.TotalAmount == 200
	})).Return(&model.PromocodeValidationResponse{
		Valid:          true,
		Code:           "SALE10",
		DiscountType:   "percent",
		DiscountValue:  10,
		OriginalAmount: 200,
		DiscountAmount: 20,
		FinalAmount:    180,
	}, nil).Once()

	w := performJSONRequest(router, http.MethodPost, "/promocodes/validate", map[string]interface{}{
		"code":         "SALE10",
		"event_id":     eventID.String(),
		"total_amount": 200,
	})

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"valid\":true")
	promocodeSvc.AssertExpectations(t)
}

func TestPromocodeHandlerGetPromocodeNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	promocodeSvc := new(promocodeServiceMock)
	h := NewPromocodeHandler(promocodeSvc)
	router := gin.New()
	router.GET("/promocodes/:id", h.GetPromocode)

	id := uuid.New()
	promocodeSvc.On("GetPromocode", mock.Anything, id).Return((*model.Promocode)(nil), repository.ErrPromocodeNotFound).Once()

	w := performJSONRequest(router, http.MethodGet, "/promocodes/"+id.String(), nil)

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "promocode not found")
	promocodeSvc.AssertExpectations(t)
}

func TestPromocodeHandlerGetAllPromocodesSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	promocodeSvc := new(promocodeServiceMock)
	h := NewPromocodeHandler(promocodeSvc)
	router := gin.New()
	router.GET("/promocodes", h.GetAllPromocodes)

	promocodeSvc.On("GetAllPromocodes", mock.Anything, 1, 20).Return([]*model.Promocode{
		{ID: uuid.New(), Code: "SALE10", DiscountType: model.DiscountTypePercent, DiscountValue: 10, IsActive: true},
	}, 1, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/promocodes", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "SALE10")
	assert.Contains(t, w.Body.String(), "\"total\":1")
	promocodeSvc.AssertExpectations(t)
}

func TestPromocodeHandlerUpdatePromocodeBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	promocodeSvc := new(promocodeServiceMock)
	h := NewPromocodeHandler(promocodeSvc)
	router := gin.New()
	router.PUT("/promocodes/:id", h.UpdatePromocode)

	id := uuid.New()
	w := performJSONRequest(router, http.MethodPut, "/promocodes/"+id.String(), map[string]interface{}{
		"discount_value": -1,
	})

	require.Equal(t, http.StatusBadRequest, w.Code)
	promocodeSvc.AssertExpectations(t)
}

func TestPromocodeHandlerDeletePromocodeSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	promocodeSvc := new(promocodeServiceMock)
	h := NewPromocodeHandler(promocodeSvc)
	router := gin.New()
	router.DELETE("/promocodes/:id", h.DeletePromocode)

	id := uuid.New()
	promocodeSvc.On("DeletePromocode", mock.Anything, id).Return(nil).Once()

	w := performJSONRequest(router, http.MethodDelete, "/promocodes/"+id.String(), nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "promocode deleted successfully")
	promocodeSvc.AssertExpectations(t)
}

func TestPromocodeHandlerDeactivatePromocodeNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	promocodeSvc := new(promocodeServiceMock)
	h := NewPromocodeHandler(promocodeSvc)
	router := gin.New()
	router.POST("/promocodes/:id/deactivate", h.DeactivatePromocode)

	id := uuid.New()
	promocodeSvc.On("DeactivatePromocode", mock.Anything, id).Return(repository.ErrPromocodeNotFound).Once()

	w := performJSONRequest(router, http.MethodPost, "/promocodes/"+id.String()+"/deactivate", nil)

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "promocode not found")
	promocodeSvc.AssertExpectations(t)
}
