package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type adminBookingServiceMock struct {
	mock.Mock
}

func (m *adminBookingServiceMock) GetAllBookings(ctx context.Context, page, limit int, status, userID, eventID string) ([]*model.Booking, int, error) {
	args := m.Called(ctx, page, limit, status, userID, eventID)
	resp, _ := args.Get(0).([]*model.Booking)
	return resp, args.Int(1), args.Error(2)
}

func (m *adminBookingServiceMock) GetBookingDetail(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) {
	args := m.Called(ctx, id)
	resp, _ := args.Get(0).(map[string]interface{})
	return resp, args.Error(1)
}

func (m *adminBookingServiceMock) CancelBooking(ctx context.Context, id uuid.UUID, refund bool) error {
	args := m.Called(ctx, id, refund)
	return args.Error(0)
}

func (m *adminBookingServiceMock) RefundBooking(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *adminBookingServiceMock) GetBookingsStats(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	resp, _ := args.Get(0).(map[string]interface{})
	return resp, args.Error(1)
}

func (m *adminBookingServiceMock) ExportBookingsToCSV(ctx context.Context, status string) ([][]string, error) {
	args := m.Called(ctx, status)
	resp, _ := args.Get(0).([][]string)
	return resp, args.Error(1)
}

func TestAdminBookingHandlerGetAllBookingsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.GET("/admin/bookings", h.GetAllBookings)

	bookingID := uuid.New()
	svc.On("GetAllBookings", mock.Anything, 1, 20, "", "", "").
		Return([]*model.Booking{{ID: bookingID, Status: model.BookingStatusPending}}, 1, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/bookings", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), bookingID.String())
	assert.Contains(t, w.Body.String(), `"total":1`)
	svc.AssertExpectations(t)
}

func TestAdminBookingHandlerGetAllBookingsWithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.GET("/admin/bookings", h.GetAllBookings)

	svc.On("GetAllBookings", mock.Anything, 2, 10, "confirmed", "", "").
		Return([]*model.Booking{}, 0, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/bookings?page=2&limit=10&status=confirmed", nil)

	require.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestAdminBookingHandlerGetAllBookingsInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.GET("/admin/bookings", h.GetAllBookings)

	svc.On("GetAllBookings", mock.Anything, 1, 20, "", "", "").
		Return(([]*model.Booking)(nil), 0, errors.New("db error")).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/bookings", nil)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to get bookings")
	svc.AssertExpectations(t)
}

func TestAdminBookingHandlerGetBookingDetailSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.GET("/admin/bookings/:id", h.GetBookingDetail)

	bookingID := uuid.New()
	svc.On("GetBookingDetail", mock.Anything, bookingID).
		Return(map[string]interface{}{"id": bookingID.String(), "status": "pending"}, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/bookings/"+bookingID.String(), nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), bookingID.String())
	svc.AssertExpectations(t)
}

func TestAdminBookingHandlerGetBookingDetailInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.GET("/admin/bookings/:id", h.GetBookingDetail)

	w := performJSONRequest(router, http.MethodGet, "/admin/bookings/not-a-uuid", nil)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid booking id")
}

func TestAdminBookingHandlerGetBookingDetailNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.GET("/admin/bookings/:id", h.GetBookingDetail)

	bookingID := uuid.New()
	svc.On("GetBookingDetail", mock.Anything, bookingID).
		Return((map[string]interface{})(nil), repository.ErrBookingNotFound).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/bookings/"+bookingID.String(), nil)

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "booking not found")
	svc.AssertExpectations(t)
}

func TestAdminBookingHandlerCancelBookingSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.POST("/admin/bookings/:id/cancel", h.CancelBooking)

	bookingID := uuid.New()
	svc.On("CancelBooking", mock.Anything, bookingID, false).Return(nil).Once()

	w := performJSONRequest(router, http.MethodPost, "/admin/bookings/"+bookingID.String()+"/cancel", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "booking cancelled successfully")
	svc.AssertExpectations(t)
}

func TestAdminBookingHandlerCancelBookingWithRefund(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.POST("/admin/bookings/:id/cancel", h.CancelBooking)

	bookingID := uuid.New()
	svc.On("CancelBooking", mock.Anything, bookingID, true).Return(nil).Once()

	w := performJSONRequest(router, http.MethodPost, "/admin/bookings/"+bookingID.String()+"/cancel?refund=true", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "booking cancelled successfully")
	svc.AssertExpectations(t)
}

func TestAdminBookingHandlerCancelBookingInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.POST("/admin/bookings/:id/cancel", h.CancelBooking)

	w := performJSONRequest(router, http.MethodPost, "/admin/bookings/bad-uuid/cancel", nil)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid booking id")
}

func TestAdminBookingHandlerCancelBookingNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.POST("/admin/bookings/:id/cancel", h.CancelBooking)

	bookingID := uuid.New()
	svc.On("CancelBooking", mock.Anything, bookingID, false).Return(repository.ErrBookingNotFound).Once()

	w := performJSONRequest(router, http.MethodPost, "/admin/bookings/"+bookingID.String()+"/cancel", nil)

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "booking not found")
	svc.AssertExpectations(t)
}

func TestAdminBookingHandlerCancelBookingCannotCancel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.POST("/admin/bookings/:id/cancel", h.CancelBooking)

	bookingID := uuid.New()
	svc.On("CancelBooking", mock.Anything, bookingID, false).Return(errors.New("cannot cancel this booking")).Once()

	w := performJSONRequest(router, http.MethodPost, "/admin/bookings/"+bookingID.String()+"/cancel", nil)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "cannot cancel this booking")
	svc.AssertExpectations(t)
}

func TestAdminBookingHandlerRefundBookingSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.POST("/admin/bookings/:id/refund", h.RefundBooking)

	bookingID := uuid.New()
	svc.On("RefundBooking", mock.Anything, bookingID).Return(nil).Once()

	w := performJSONRequest(router, http.MethodPost, "/admin/bookings/"+bookingID.String()+"/refund", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "refund processed successfully")
	svc.AssertExpectations(t)
}

func TestAdminBookingHandlerRefundBookingInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.POST("/admin/bookings/:id/refund", h.RefundBooking)

	w := performJSONRequest(router, http.MethodPost, "/admin/bookings/bad-uuid/refund", nil)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid booking id")
}

func TestAdminBookingHandlerRefundBookingNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.POST("/admin/bookings/:id/refund", h.RefundBooking)

	bookingID := uuid.New()
	svc.On("RefundBooking", mock.Anything, bookingID).Return(repository.ErrBookingNotFound).Once()

	w := performJSONRequest(router, http.MethodPost, "/admin/bookings/"+bookingID.String()+"/refund", nil)

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "booking not found")
	svc.AssertExpectations(t)
}

func TestAdminBookingHandlerRefundBookingPaymentNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.POST("/admin/bookings/:id/refund", h.RefundBooking)

	bookingID := uuid.New()
	svc.On("RefundBooking", mock.Anything, bookingID).Return(repository.ErrPaymentNotFound).Once()

	w := performJSONRequest(router, http.MethodPost, "/admin/bookings/"+bookingID.String()+"/refund", nil)

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "payment not found")
	svc.AssertExpectations(t)
}

func TestAdminBookingHandlerRefundBookingAlreadyRefunded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.POST("/admin/bookings/:id/refund", h.RefundBooking)

	bookingID := uuid.New()
	svc.On("RefundBooking", mock.Anything, bookingID).Return(errors.New("payment already refunded")).Once()

	w := performJSONRequest(router, http.MethodPost, "/admin/bookings/"+bookingID.String()+"/refund", nil)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "payment already refunded")
	svc.AssertExpectations(t)
}

func TestAdminBookingHandlerGetBookingsStatsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.GET("/admin/bookings/stats", h.GetBookingsStats)

	svc.On("GetBookingsStats", mock.Anything).
		Return(map[string]interface{}{"total": 42, "cancelled": 5}, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/bookings/stats", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"total\"")
	svc.AssertExpectations(t)
}

func TestAdminBookingHandlerGetBookingsStatsInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.GET("/admin/bookings/stats", h.GetBookingsStats)

	svc.On("GetBookingsStats", mock.Anything).
		Return((map[string]interface{})(nil), errors.New("db error")).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/bookings/stats", nil)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to get stats")
	svc.AssertExpectations(t)
}

func TestAdminBookingHandlerExportBookingsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.GET("/admin/bookings/export", h.ExportBookings)

	svc.On("ExportBookingsToCSV", mock.Anything, "").
		Return([][]string{
			{"id", "status", "amount"},
			{"some-uuid", "confirmed", "100"},
		}, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/bookings/export", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "bookings_export.csv")
	assert.Contains(t, w.Body.String(), "confirmed")
	svc.AssertExpectations(t)
}

func TestAdminBookingHandlerExportBookingsWithStatusFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.GET("/admin/bookings/export", h.ExportBookings)

	svc.On("ExportBookingsToCSV", mock.Anything, "cancelled").
		Return([][]string{{"id", "status"}, {"uuid-1", "cancelled"}}, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/bookings/export?status=cancelled", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "cancelled")
	svc.AssertExpectations(t)
}

func TestAdminBookingHandlerExportBookingsInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminBookingServiceMock)
	h := NewAdminBookingHandler(svc)
	router := gin.New()
	router.GET("/admin/bookings/export", h.ExportBookings)

	svc.On("ExportBookingsToCSV", mock.Anything, "").
		Return(([][]string)(nil), errors.New("export failed")).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/bookings/export", nil)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to export bookings")
	svc.AssertExpectations(t)
}
