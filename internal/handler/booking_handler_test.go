package handler

import (
	"context"
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

type bookingServiceMock struct {
	mock.Mock
}

func (m *bookingServiceMock) CreateBooking(ctx context.Context, userID uuid.UUID, req *model.CreateBookingRequest) (*model.BookingResponse, error) {
	args := m.Called(ctx, userID, req)
	resp, _ := args.Get(0).(*model.BookingResponse)
	return resp, args.Error(1)
}

func (m *bookingServiceMock) GetBooking(ctx context.Context, id uuid.UUID) (*model.BookingResponse, error) {
	args := m.Called(ctx, id)
	resp, _ := args.Get(0).(*model.BookingResponse)
	return resp, args.Error(1)
}

func (m *bookingServiceMock) GetUserBookings(ctx context.Context, userID uuid.UUID, page, limit int) ([]*model.BookingResponse, int, error) {
	args := m.Called(ctx, userID, page, limit)
	resp, _ := args.Get(0).([]*model.BookingResponse)
	return resp, args.Int(1), args.Error(2)
}

func (m *bookingServiceMock) CancelBooking(ctx context.Context, bookingID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, bookingID, userID)
	return args.Error(0)
}

func (m *bookingServiceMock) ConfirmBooking(ctx context.Context, bookingID uuid.UUID) error {
	args := m.Called(ctx, bookingID)
	return args.Error(0)
}

func (m *bookingServiceMock) ExpirePendingBookings(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func TestBookingHandlerCreateBookingSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	bookingSvc := new(bookingServiceMock)
	h := NewBookingHandler(bookingSvc)
	router := gin.New()
	router.Use(addUserContext())
	router.POST("/bookings", h.CreateBooking)

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	eventID := uuid.New()
	seatID := uuid.New()

	bookingSvc.On("CreateBooking", mock.Anything, userID, mock.MatchedBy(func(req *model.CreateBookingRequest) bool {
		return req.EventID == eventID && len(req.SeatIDs) == 1 && req.SeatIDs[0] == seatID
	})).Return(&model.BookingResponse{
		ID:          uuid.New(),
		UserID:      userID,
		EventID:     eventID,
		FinalAmount: 100,
		Status:      model.BookingStatusPending,
		ExpiresAt:   time.Now().Add(15 * time.Minute),
	}, nil).Once()

	w := performJSONRequest(router, http.MethodPost, "/bookings", map[string]interface{}{
		"event_id": eventID.String(),
		"seat_ids": []string{seatID.String()},
	})

	require.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), eventID.String())
	bookingSvc.AssertExpectations(t)
}

func TestBookingHandlerGetBookingForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	bookingSvc := new(bookingServiceMock)
	h := NewBookingHandler(bookingSvc)
	router := gin.New()
	router.Use(addUserContext())
	router.GET("/bookings/:id", h.GetBooking)

	bookingID := uuid.New()
	bookingSvc.On("GetBooking", mock.Anything, bookingID).Return(&model.BookingResponse{
		ID:     bookingID,
		UserID: uuid.New(),
	}, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/bookings/"+bookingID.String(), nil)

	require.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized to view this booking")
	bookingSvc.AssertExpectations(t)
}

func TestBookingHandlerGetUserBookingsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	bookingSvc := new(bookingServiceMock)
	h := NewBookingHandler(bookingSvc)
	router := gin.New()
	router.Use(addUserContext())
	router.GET("/bookings", h.GetUserBookings)

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	bookingSvc.On("GetUserBookings", mock.Anything, userID, 2, 10).Return([]*model.BookingResponse{
		{ID: uuid.New(), UserID: userID},
	}, 1, nil).Once()

	req := performJSONRequest(router, http.MethodGet, "/bookings?page=2&limit=10", nil)

	require.Equal(t, http.StatusOK, req.Code)
	assert.Contains(t, req.Body.String(), "\"total\":1")
	bookingSvc.AssertExpectations(t)
}

func TestBookingHandlerCancelBookingCannotCancel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	bookingSvc := new(bookingServiceMock)
	h := NewBookingHandler(bookingSvc)
	router := gin.New()
	router.Use(addUserContext())
	router.POST("/bookings/:id/cancel", h.CancelBooking)

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	bookingID := uuid.New()
	bookingSvc.On("CancelBooking", mock.Anything, bookingID, userID).Return(service.ErrCannotCancel).Once()

	w := performJSONRequest(router, http.MethodPost, "/bookings/"+bookingID.String()+"/cancel", nil)

	require.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), service.ErrCannotCancel.Error())
	bookingSvc.AssertExpectations(t)
}

func TestBookingHandlerConfirmBookingExpired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	bookingSvc := new(bookingServiceMock)
	h := NewBookingHandler(bookingSvc)
	router := gin.New()
	router.POST("/bookings/:id/confirm", h.ConfirmBooking)

	bookingID := uuid.New()
	bookingSvc.On("ConfirmBooking", mock.Anything, bookingID).Return(service.ErrBookingExpired).Once()

	w := performJSONRequest(router, http.MethodPost, "/bookings/"+bookingID.String()+"/confirm", nil)

	require.Equal(t, http.StatusGone, w.Code)
	assert.Contains(t, w.Body.String(), service.ErrBookingExpired.Error())
	bookingSvc.AssertExpectations(t)
}

func TestBookingHandlerCreateBookingNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	bookingSvc := new(bookingServiceMock)
	h := NewBookingHandler(bookingSvc)
	router := gin.New()
	router.Use(addUserContext())
	router.POST("/bookings", h.CreateBooking)

	eventID := uuid.New()
	seatID := uuid.New()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	bookingSvc.On("CreateBooking", mock.Anything, userID, mock.AnythingOfType("*model.CreateBookingRequest")).
		Return((*model.BookingResponse)(nil), repository.ErrSeatNotFound).Once()

	w := performJSONRequest(router, http.MethodPost, "/bookings", map[string]interface{}{
		"event_id": eventID.String(),
		"seat_ids": []string{seatID.String()},
	})

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), repository.ErrSeatNotFound.Error())
	bookingSvc.AssertExpectations(t)
}
