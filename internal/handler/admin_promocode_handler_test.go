package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/bekzat-kamen/booking_system_api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type adminPromocodeServiceMock struct {
	mock.Mock
}

func (m *adminPromocodeServiceMock) GetAllPromocodes(ctx context.Context, page, limit int, isActive string) ([]*model.Promocode, int, error) {
	args := m.Called(ctx, page, limit, isActive)
	resp, _ := args.Get(0).([]*model.Promocode)
	return resp, args.Int(1), args.Error(2)
}

func (m *adminPromocodeServiceMock) GetPromocodeDetail(ctx context.Context, id uuid.UUID) (*model.PromocodeDetail, error) {
	args := m.Called(ctx, id)
	resp, _ := args.Get(0).(*model.PromocodeDetail)
	return resp, args.Error(1)
}

func (m *adminPromocodeServiceMock) UpdatePromocode(ctx context.Context, id uuid.UUID, req *model.UpdatePromocodeRequest) (*model.Promocode, error) {
	args := m.Called(ctx, id, req)
	resp, _ := args.Get(0).(*model.Promocode)
	return resp, args.Error(1)
}

func (m *adminPromocodeServiceMock) DeletePromocode(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *adminPromocodeServiceMock) BulkDeactivate(ctx context.Context, ids []uuid.UUID) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *adminPromocodeServiceMock) GetPromocodesStats(ctx context.Context) (*model.PromocodesStats, error) {
	args := m.Called(ctx)
	resp, _ := args.Get(0).(*model.PromocodesStats)
	return resp, args.Error(1)
}

func (m *adminPromocodeServiceMock) ExportPromocodesToCSV(ctx context.Context) ([][]string, error) {
	args := m.Called(ctx)
	resp, _ := args.Get(0).([][]string)
	return resp, args.Error(1)
}

func newTestPromocode(id uuid.UUID) *model.Promocode {
	return &model.Promocode{
		ID:            id,
		Code:          "TEST10",
		Description:   "Test Promocode",
		DiscountType:  model.DiscountTypePercent,
		DiscountValue: 10,
		IsActive:      true,
		ValidFrom:     time.Now(),
	}
}

func TestAdminPromocodeHandlerGetAllPromocodesSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.GET("/admin/promocodes", h.GetAllPromocodes)

	promoID := uuid.New()
	svc.On("GetAllPromocodes", mock.Anything, 1, 20, "").
		Return([]*model.Promocode{newTestPromocode(promoID)}, 1, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/promocodes", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), promoID.String())
	assert.Contains(t, w.Body.String(), `"total":1`)
	svc.AssertExpectations(t)
}

func TestAdminPromocodeHandlerGetAllPromocodesWithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.GET("/admin/promocodes", h.GetAllPromocodes)

	svc.On("GetAllPromocodes", mock.Anything, 2, 10, "true").
		Return([]*model.Promocode{}, 0, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/promocodes?page=2&limit=10&is_active=true", nil)

	require.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestAdminPromocodeHandlerGetAllPromocodesInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.GET("/admin/promocodes", h.GetAllPromocodes)

	svc.On("GetAllPromocodes", mock.Anything, 1, 20, "").
		Return(([]*model.Promocode)(nil), 0, errors.New("db error")).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/promocodes", nil)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to get promocodes")
	svc.AssertExpectations(t)
}

func TestAdminPromocodeHandlerGetPromocodeDetailSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.GET("/admin/promocodes/:id", h.GetPromocodeDetail)

	promoID := uuid.New()
	svc.On("GetPromocodeDetail", mock.Anything, promoID).
		Return(&model.PromocodeDetail{
			Promocode: model.Promocode{ID: promoID, Code: "TEST10"},
		}, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/promocodes/"+promoID.String(), nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), promoID.String())
	assert.Contains(t, w.Body.String(), "TEST10")
	svc.AssertExpectations(t)
}

func TestAdminPromocodeHandlerGetPromocodeDetailInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.GET("/admin/promocodes/:id", h.GetPromocodeDetail)

	w := performJSONRequest(router, http.MethodGet, "/admin/promocodes/not-a-uuid", nil)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid promocode id")
}

func TestAdminPromocodeHandlerGetPromocodeDetailNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.GET("/admin/promocodes/:id", h.GetPromocodeDetail)

	promoID := uuid.New()
	svc.On("GetPromocodeDetail", mock.Anything, promoID).
		Return((*model.PromocodeDetail)(nil), repository.ErrPromocodeNotFound).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/promocodes/"+promoID.String(), nil)

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "promocode not found")
	svc.AssertExpectations(t)
}

func TestAdminPromocodeHandlerGetPromocodeDetailInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.GET("/admin/promocodes/:id", h.GetPromocodeDetail)

	promoID := uuid.New()
	svc.On("GetPromocodeDetail", mock.Anything, promoID).
		Return((*model.PromocodeDetail)(nil), errors.New("db error")).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/promocodes/"+promoID.String(), nil)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to get promocode")
	svc.AssertExpectations(t)
}

func TestAdminPromocodeHandlerUpdatePromocodeSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.PUT("/admin/promocodes/:id", h.UpdatePromocode)

	promoID := uuid.New()
	svc.On("UpdatePromocode", mock.Anything, promoID, mock.AnythingOfType("*model.UpdatePromocodeRequest")).
		Return(newTestPromocode(promoID), nil).Once()

	w := performJSONRequest(router, http.MethodPut, "/admin/promocodes/"+promoID.String(), map[string]interface{}{
		"description": "Updated",
	})

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), promoID.String())
	svc.AssertExpectations(t)
}

func TestAdminPromocodeHandlerUpdatePromocodeInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.PUT("/admin/promocodes/:id", h.UpdatePromocode)

	w := performJSONRequest(router, http.MethodPut, "/admin/promocodes/bad-uuid", map[string]interface{}{})

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid promocode id")
}

func TestAdminPromocodeHandlerUpdatePromocodeNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.PUT("/admin/promocodes/:id", h.UpdatePromocode)

	promoID := uuid.New()
	svc.On("UpdatePromocode", mock.Anything, promoID, mock.AnythingOfType("*model.UpdatePromocodeRequest")).
		Return((*model.Promocode)(nil), repository.ErrPromocodeNotFound).Once()

	w := performJSONRequest(router, http.MethodPut, "/admin/promocodes/"+promoID.String(), map[string]interface{}{})

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "promocode not found")
	svc.AssertExpectations(t)
}

func TestAdminPromocodeHandlerUpdatePromocodeInvalidDiscount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.PUT("/admin/promocodes/:id", h.UpdatePromocode)

	promoID := uuid.New()
	svc.On("UpdatePromocode", mock.Anything, promoID, mock.AnythingOfType("*model.UpdatePromocodeRequest")).
		Return((*model.Promocode)(nil), service.ErrInvalidDiscountValue).Once()

	w := performJSONRequest(router, http.MethodPut, "/admin/promocodes/"+promoID.String(), map[string]interface{}{})

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), service.ErrInvalidDiscountValue.Error())
	svc.AssertExpectations(t)
}

func TestAdminPromocodeHandlerUpdatePromocodeInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.PUT("/admin/promocodes/:id", h.UpdatePromocode)

	promoID := uuid.New()
	svc.On("UpdatePromocode", mock.Anything, promoID, mock.AnythingOfType("*model.UpdatePromocodeRequest")).
		Return((*model.Promocode)(nil), errors.New("db error")).Once()

	w := performJSONRequest(router, http.MethodPut, "/admin/promocodes/"+promoID.String(), map[string]interface{}{})

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to update promocode")
	svc.AssertExpectations(t)
}

func TestAdminPromocodeHandlerDeletePromocodeSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.DELETE("/admin/promocodes/:id", h.DeletePromocode)

	promoID := uuid.New()
	svc.On("DeletePromocode", mock.Anything, promoID).Return(nil).Once()

	w := performJSONRequest(router, http.MethodDelete, "/admin/promocodes/"+promoID.String(), nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "promocode deleted successfully")
	svc.AssertExpectations(t)
}

func TestAdminPromocodeHandlerDeletePromocodeInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.DELETE("/admin/promocodes/:id", h.DeletePromocode)

	w := performJSONRequest(router, http.MethodDelete, "/admin/promocodes/bad-uuid", nil)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid promocode id")
}

func TestAdminPromocodeHandlerDeletePromocodeNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.DELETE("/admin/promocodes/:id", h.DeletePromocode)

	promoID := uuid.New()
	svc.On("DeletePromocode", mock.Anything, promoID).Return(repository.ErrPromocodeNotFound).Once()

	w := performJSONRequest(router, http.MethodDelete, "/admin/promocodes/"+promoID.String(), nil)

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "promocode not found")
	svc.AssertExpectations(t)
}

func TestAdminPromocodeHandlerDeletePromocodeInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.DELETE("/admin/promocodes/:id", h.DeletePromocode)

	promoID := uuid.New()
	svc.On("DeletePromocode", mock.Anything, promoID).Return(errors.New("db error")).Once()

	w := performJSONRequest(router, http.MethodDelete, "/admin/promocodes/"+promoID.String(), nil)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to delete promocode")
	svc.AssertExpectations(t)
}

func TestAdminPromocodeHandlerBulkDeactivateSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.POST("/admin/promocodes/bulk-deactivate", h.BulkDeactivate)

	id1 := uuid.New()
	id2 := uuid.New()
	ids := []uuid.UUID{id1, id2}
	svc.On("BulkDeactivate", mock.Anything, ids).Return(nil).Once()

	w := performJSONRequest(router, http.MethodPost, "/admin/promocodes/bulk-deactivate", map[string]interface{}{
		"ids": []string{id1.String(), id2.String()},
	})

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "promocodes deactivated successfully")
	svc.AssertExpectations(t)
}

func TestAdminPromocodeHandlerBulkDeactivateInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.POST("/admin/promocodes/bulk-deactivate", h.BulkDeactivate)

	w := performJSONRequest(router, http.MethodPost, "/admin/promocodes/bulk-deactivate", map[string]interface{}{
		"ids": []string{"not-a-uuid"},
	})

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid uuid")
}

func TestAdminPromocodeHandlerBulkDeactivateInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.POST("/admin/promocodes/bulk-deactivate", h.BulkDeactivate)

	id1 := uuid.New()
	svc.On("BulkDeactivate", mock.Anything, []uuid.UUID{id1}).Return(errors.New("db error")).Once()

	w := performJSONRequest(router, http.MethodPost, "/admin/promocodes/bulk-deactivate", map[string]interface{}{
		"ids": []string{id1.String()},
	})

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to deactivate promocodes")
	svc.AssertExpectations(t)
}

func TestAdminPromocodeHandlerGetPromocodesStatsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.GET("/admin/promocodes/stats", h.GetPromocodesStats)

	svc.On("GetPromocodesStats", mock.Anything).
		Return(&model.PromocodesStats{TotalPromocodes: 10}, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/promocodes/stats", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"total_promocodes\":10")
	svc.AssertExpectations(t)
}

func TestAdminPromocodeHandlerGetPromocodesStatsInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.GET("/admin/promocodes/stats", h.GetPromocodesStats)

	svc.On("GetPromocodesStats", mock.Anything).
		Return((*model.PromocodesStats)(nil), errors.New("db error")).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/promocodes/stats", nil)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to get stats")
	svc.AssertExpectations(t)
}

func TestAdminPromocodeHandlerExportPromocodesSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.GET("/admin/promocodes/export", h.ExportPromocodes)

	svc.On("ExportPromocodesToCSV", mock.Anything).
		Return([][]string{
			{"code", "discount"},
			{"TEST10", "10"},
		}, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/promocodes/export", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "promocodes_export.csv")
	assert.Contains(t, w.Body.String(), "TEST10")
	svc.AssertExpectations(t)
}

func TestAdminPromocodeHandlerExportPromocodesInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminPromocodeServiceMock)
	h := NewAdminPromocodeHandler(svc)
	router := gin.New()
	router.GET("/admin/promocodes/export", h.ExportPromocodes)

	svc.On("ExportPromocodesToCSV", mock.Anything).
		Return(([][]string)(nil), errors.New("export failed")).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/promocodes/export", nil)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to export promocodes")
	svc.AssertExpectations(t)
}
