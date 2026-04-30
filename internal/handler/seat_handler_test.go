package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/bekzat-kamen/booking_system_api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type seatServiceMock struct {
	mock.Mock
}

func (m *seatServiceMock) GenerateSeats(ctx context.Context, eventID, organizerID uuid.UUID, req *model.GenerateSeatsRequest) error {
	args := m.Called(ctx, eventID, organizerID, req)
	return args.Error(0)
}

func (m *seatServiceMock) GetSeatMap(ctx context.Context, eventID uuid.UUID) (*model.SeatMapResponse, error) {
	args := m.Called(ctx, eventID)
	resp, _ := args.Get(0).(*model.SeatMapResponse)
	return resp, args.Error(1)
}

func (m *seatServiceMock) GetAvailableSeats(ctx context.Context, eventID uuid.UUID) ([]model.SeatResponse, error) {
	args := m.Called(ctx, eventID)
	resp, _ := args.Get(0).([]model.SeatResponse)
	return resp, args.Error(1)
}

func (m *seatServiceMock) ReserveSeat(ctx context.Context, seatID uuid.UUID) error {
	args := m.Called(ctx, seatID)
	return args.Error(0)
}

func (m *seatServiceMock) BookSeat(ctx context.Context, seatID uuid.UUID) error {
	args := m.Called(ctx, seatID)
	return args.Error(0)
}

func (m *seatServiceMock) ReleaseSeat(ctx context.Context, seatID uuid.UUID) error {
	args := m.Called(ctx, seatID)
	return args.Error(0)
}

func TestSeatHandlerGenerateSeatsConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(seatServiceMock)
	h := NewSeatHandler(svc)
	router := gin.New()
	router.Use(addUserContext())
	router.POST("/events/:id/seats/generate", h.GenerateSeats)

	eventID := uuid.New()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	svc.On("GenerateSeats", mock.Anything, eventID, userID, mock.AnythingOfType("*model.GenerateSeatsRequest")).
		Return(service.ErrSeatsAlreadyGenerated).Once()

	w := performJSONRequest(router, http.MethodPost, "/events/"+eventID.String()+"/seats/generate", map[string]interface{}{
		"total_rows":    2,
		"seats_per_row": 3,
		"base_price":    10,
	})

	require.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "seats already generated for this event")
}

func TestSeatHandlerGetSeatMapSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(seatServiceMock)
	h := NewSeatHandler(svc)
	router := gin.New()
	router.GET("/events/:id/seats", h.GetSeatMap)

	eventID := uuid.New()
	svc.On("GetSeatMap", mock.Anything, eventID).Return(&model.SeatMapResponse{EventID: eventID, TotalSeats: 2}, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/events/"+eventID.String()+"/seats", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), eventID.String())
}

func TestSeatHandlerGetSeatMapNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(seatServiceMock)
	h := NewSeatHandler(svc)
	router := gin.New()
	router.GET("/events/:id/seats", h.GetSeatMap)

	eventID := uuid.New()
	svc.On("GetSeatMap", mock.Anything, eventID).Return((*model.SeatMapResponse)(nil), repository.ErrEventNotFound).Once()

	w := performJSONRequest(router, http.MethodGet, "/events/"+eventID.String()+"/seats", nil)

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "event not found")
}

func TestSeatHandlerGetAvailableSeatsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(seatServiceMock)
	h := NewSeatHandler(svc)
	router := gin.New()
	router.GET("/events/:id/seats/available", h.GetAvailableSeats)

	eventID := uuid.New()
	svc.On("GetAvailableSeats", mock.Anything, eventID).Return([]model.SeatResponse{{ID: uuid.New(), EventID: eventID}}, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/events/"+eventID.String()+"/seats/available", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"seats\"")
}
