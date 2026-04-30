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

type eventServiceMock struct {
	mock.Mock
}

func (m *eventServiceMock) Create(ctx context.Context, organizerID uuid.UUID, req *model.CreateEventRequest) (*model.Event, error) {
	args := m.Called(ctx, organizerID, req)
	resp, _ := args.Get(0).(*model.Event)
	return resp, args.Error(1)
}

func (m *eventServiceMock) GetByID(ctx context.Context, eventID uuid.UUID) (*model.Event, error) {
	args := m.Called(ctx, eventID)
	resp, _ := args.Get(0).(*model.Event)
	return resp, args.Error(1)
}

func (m *eventServiceMock) GetAll(ctx context.Context, limit, offset int) ([]*model.Event, int, error) {
	args := m.Called(ctx, limit, offset)
	resp, _ := args.Get(0).([]*model.Event)
	return resp, args.Int(1), args.Error(2)
}

func (m *eventServiceMock) GetByOrganizer(ctx context.Context, organizerID uuid.UUID, limit, offset int) ([]*model.Event, error) {
	args := m.Called(ctx, organizerID, limit, offset)
	resp, _ := args.Get(0).([]*model.Event)
	return resp, args.Error(1)
}

func (m *eventServiceMock) Update(ctx context.Context, eventID, organizerID uuid.UUID, req *model.UpdateEventRequest) (*model.Event, error) {
	args := m.Called(ctx, eventID, organizerID, req)
	resp, _ := args.Get(0).(*model.Event)
	return resp, args.Error(1)
}

func (m *eventServiceMock) Delete(ctx context.Context, eventID, organizerID uuid.UUID) error {
	args := m.Called(ctx, eventID, organizerID)
	return args.Error(0)
}

func (m *eventServiceMock) PublishEvent(ctx context.Context, eventID, organizerID uuid.UUID) (*model.Event, error) {
	args := m.Called(ctx, eventID, organizerID)
	resp, _ := args.Get(0).(*model.Event)
	return resp, args.Error(1)
}

func TestEventHandlerCreateSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(eventServiceMock)
	h := NewEventHandler(svc)
	router := gin.New()
	router.Use(addUserContext())
	router.POST("/events", h.Create)

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	eventDate := time.Now().Add(time.Hour).Format(time.RFC3339)
	svc.On("Create", mock.Anything, userID, mock.MatchedBy(func(req *model.CreateEventRequest) bool {
		return req.Title == "Concert" && req.EventDate == eventDate
	})).Return(&model.Event{ID: uuid.New(), Title: "Concert"}, nil).Once()

	w := performJSONRequest(router, http.MethodPost, "/events", map[string]interface{}{
		"title":       "Concert",
		"venue":       "Hall",
		"event_date":  eventDate,
		"total_seats": 10,
		"price":       100,
	})

	require.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Concert")
}

func TestEventHandlerGetByIDNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(eventServiceMock)
	h := NewEventHandler(svc)
	router := gin.New()
	router.GET("/events/:id", h.GetByID)

	eventID := uuid.New()
	svc.On("GetByID", mock.Anything, eventID).Return((*model.Event)(nil), repository.ErrEventNotFound).Once()

	w := performJSONRequest(router, http.MethodGet, "/events/"+eventID.String(), nil)

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "event not found")
}

func TestEventHandlerGetAllSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(eventServiceMock)
	h := NewEventHandler(svc)
	router := gin.New()
	router.GET("/events", h.GetAll)

	svc.On("GetAll", mock.Anything, 5, 5).Return([]*model.Event{{ID: uuid.New(), Title: "Concert"}}, 1, nil).Once()

	req := performJSONRequest(router, http.MethodGet, "/events?limit=5&page=2", nil)

	require.Equal(t, http.StatusOK, req.Code)
	assert.Contains(t, req.Body.String(), "\"total\":1")
}

func TestEventHandlerUpdateForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(eventServiceMock)
	h := NewEventHandler(svc)
	router := gin.New()
	router.Use(addUserContext())
	router.PUT("/events/:id", h.Update)

	eventID := uuid.New()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	svc.On("Update", mock.Anything, eventID, userID, mock.AnythingOfType("*model.UpdateEventRequest")).
		Return((*model.Event)(nil), service.ErrEventForbidden).Once()

	w := performJSONRequest(router, http.MethodPut, "/events/"+eventID.String(), map[string]interface{}{
		"title": "New title",
	})

	require.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "you are not allowed to modify this event")
}

func TestEventHandlerPublishEventSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(eventServiceMock)
	h := NewEventHandler(svc)
	router := gin.New()
	router.Use(addUserContext())
	router.POST("/events/:id/publish", h.PublishEvent)

	eventID := uuid.New()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	svc.On("PublishEvent", mock.Anything, eventID, userID).Return(&model.Event{ID: eventID, Status: model.EventStatusPublished}, nil).Once()

	w := performJSONRequest(router, http.MethodPost, "/events/"+eventID.String()+"/publish", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "published")
}
