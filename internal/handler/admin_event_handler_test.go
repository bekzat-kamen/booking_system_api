package handler

import (
	"context"
	"errors"
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

type adminEventServiceMock struct {
	mock.Mock
}

func (m *adminEventServiceMock) GetAllEvents(ctx context.Context, page, limit int, status, organizerID string) ([]*model.Event, int, error) {
	args := m.Called(ctx, page, limit, status, organizerID)
	resp, _ := args.Get(0).([]*model.Event)
	return resp, args.Int(1), args.Error(2)
}

func (m *adminEventServiceMock) GetEventDetail(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) {
	args := m.Called(ctx, id)
	resp, _ := args.Get(0).(map[string]interface{})
	return resp, args.Error(1)
}

func (m *adminEventServiceMock) UpdateEvent(ctx context.Context, id uuid.UUID, req *model.UpdateEventRequest) (*model.Event, error) {
	args := m.Called(ctx, id, req)
	resp, _ := args.Get(0).(*model.Event)
	return resp, args.Error(1)
}

func (m *adminEventServiceMock) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *adminEventServiceMock) PublishEvent(ctx context.Context, id uuid.UUID) (*model.Event, error) {
	args := m.Called(ctx, id)
	resp, _ := args.Get(0).(*model.Event)
	return resp, args.Error(1)
}

func (m *adminEventServiceMock) GetEventsStats(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	resp, _ := args.Get(0).(map[string]int64)
	return resp, args.Error(1)
}

func newTestEvent(id uuid.UUID) *model.Event {
	return &model.Event{
		ID:        id,
		Title:     "Test Event",
		Venue:     "Test Venue",
		EventDate: time.Now().Add(24 * time.Hour),
		Status:    model.EventStatusDraft,
	}
}

func TestAdminEventHandlerGetAllEventsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.GET("/admin/events", h.GetAllEvents)

	eventID := uuid.New()
	svc.On("GetAllEvents", mock.Anything, 1, 20, "", "").
		Return([]*model.Event{newTestEvent(eventID)}, 1, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/events", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), eventID.String())
	assert.Contains(t, w.Body.String(), `"total":1`)
	svc.AssertExpectations(t)
}

func TestAdminEventHandlerGetAllEventsWithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.GET("/admin/events", h.GetAllEvents)

	organizerID := uuid.New().String()
	svc.On("GetAllEvents", mock.Anything, 2, 10, "published", organizerID).
		Return([]*model.Event{}, 0, nil).Once()

	w := performJSONRequest(router, http.MethodGet,
		"/admin/events?page=2&limit=10&status=published&organizer_id="+organizerID, nil)

	require.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestAdminEventHandlerGetAllEventsInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.GET("/admin/events", h.GetAllEvents)

	svc.On("GetAllEvents", mock.Anything, 1, 20, "", "").
		Return(([]*model.Event)(nil), 0, errors.New("db error")).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/events", nil)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to get events")
	svc.AssertExpectations(t)
}

func TestAdminEventHandlerGetEventDetailSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.GET("/admin/events/:id", h.GetEventDetail)

	eventID := uuid.New()
	svc.On("GetEventDetail", mock.Anything, eventID).
		Return(map[string]interface{}{"id": eventID.String(), "title": "Test Event"}, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/events/"+eventID.String(), nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), eventID.String())
	assert.Contains(t, w.Body.String(), "Test Event")
	svc.AssertExpectations(t)
}

func TestAdminEventHandlerGetEventDetailInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.GET("/admin/events/:id", h.GetEventDetail)

	w := performJSONRequest(router, http.MethodGet, "/admin/events/not-a-uuid", nil)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid event id")
}

func TestAdminEventHandlerGetEventDetailNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.GET("/admin/events/:id", h.GetEventDetail)

	eventID := uuid.New()
	svc.On("GetEventDetail", mock.Anything, eventID).
		Return((map[string]interface{})(nil), repository.ErrEventNotFound).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/events/"+eventID.String(), nil)

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "event not found")
	svc.AssertExpectations(t)
}

func TestAdminEventHandlerGetEventDetailInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.GET("/admin/events/:id", h.GetEventDetail)

	eventID := uuid.New()
	svc.On("GetEventDetail", mock.Anything, eventID).
		Return((map[string]interface{})(nil), errors.New("db error")).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/events/"+eventID.String(), nil)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to get event")
	svc.AssertExpectations(t)
}

func TestAdminEventHandlerUpdateEventSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.PUT("/admin/events/:id", h.UpdateEvent)

	eventID := uuid.New()
	svc.On("UpdateEvent", mock.Anything, eventID, mock.AnythingOfType("*model.UpdateEventRequest")).
		Return(newTestEvent(eventID), nil).Once()

	w := performJSONRequest(router, http.MethodPut, "/admin/events/"+eventID.String(), map[string]interface{}{
		"title": "Updated Title",
		"venue": "Updated Venue",
	})

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), eventID.String())
	svc.AssertExpectations(t)
}

func TestAdminEventHandlerUpdateEventInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.PUT("/admin/events/:id", h.UpdateEvent)

	w := performJSONRequest(router, http.MethodPut, "/admin/events/bad-uuid", map[string]interface{}{
		"title": "Updated Title",
	})

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid event id")
}

func TestAdminEventHandlerUpdateEventNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.PUT("/admin/events/:id", h.UpdateEvent)

	eventID := uuid.New()
	svc.On("UpdateEvent", mock.Anything, eventID, mock.AnythingOfType("*model.UpdateEventRequest")).
		Return((*model.Event)(nil), repository.ErrEventNotFound).Once()

	w := performJSONRequest(router, http.MethodPut, "/admin/events/"+eventID.String(), map[string]interface{}{
		"title": "Updated Title",
	})

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "event not found")
	svc.AssertExpectations(t)
}

func TestAdminEventHandlerUpdateEventInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.PUT("/admin/events/:id", h.UpdateEvent)

	eventID := uuid.New()
	svc.On("UpdateEvent", mock.Anything, eventID, mock.AnythingOfType("*model.UpdateEventRequest")).
		Return((*model.Event)(nil), errors.New("db error")).Once()

	w := performJSONRequest(router, http.MethodPut, "/admin/events/"+eventID.String(), map[string]interface{}{
		"title": "Updated Title",
	})

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to update event")
	svc.AssertExpectations(t)
}

func TestAdminEventHandlerDeleteEventSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.DELETE("/admin/events/:id", h.DeleteEvent)

	eventID := uuid.New()
	svc.On("DeleteEvent", mock.Anything, eventID).Return(nil).Once()

	w := performJSONRequest(router, http.MethodDelete, "/admin/events/"+eventID.String(), nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "event cancelled successfully")
	svc.AssertExpectations(t)
}

func TestAdminEventHandlerDeleteEventInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.DELETE("/admin/events/:id", h.DeleteEvent)

	w := performJSONRequest(router, http.MethodDelete, "/admin/events/bad-uuid", nil)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid event id")
}

func TestAdminEventHandlerDeleteEventNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.DELETE("/admin/events/:id", h.DeleteEvent)

	eventID := uuid.New()
	svc.On("DeleteEvent", mock.Anything, eventID).Return(repository.ErrEventNotFound).Once()

	w := performJSONRequest(router, http.MethodDelete, "/admin/events/"+eventID.String(), nil)

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "event not found")
	svc.AssertExpectations(t)
}

func TestAdminEventHandlerDeleteEventInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.DELETE("/admin/events/:id", h.DeleteEvent)

	eventID := uuid.New()
	svc.On("DeleteEvent", mock.Anything, eventID).Return(errors.New("db error")).Once()

	w := performJSONRequest(router, http.MethodDelete, "/admin/events/"+eventID.String(), nil)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to cancel event")
	svc.AssertExpectations(t)
}

func TestAdminEventHandlerPublishEventSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.POST("/admin/events/:id/publish", h.PublishEvent)

	eventID := uuid.New()
	published := newTestEvent(eventID)
	published.Status = model.EventStatusPublished
	svc.On("PublishEvent", mock.Anything, eventID).Return(published, nil).Once()

	w := performJSONRequest(router, http.MethodPost, "/admin/events/"+eventID.String()+"/publish", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "published")
	svc.AssertExpectations(t)
}

func TestAdminEventHandlerPublishEventInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.POST("/admin/events/:id/publish", h.PublishEvent)

	w := performJSONRequest(router, http.MethodPost, "/admin/events/bad-uuid/publish", nil)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid event id")
}

func TestAdminEventHandlerPublishEventNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.POST("/admin/events/:id/publish", h.PublishEvent)

	eventID := uuid.New()
	svc.On("PublishEvent", mock.Anything, eventID).Return((*model.Event)(nil), repository.ErrEventNotFound).Once()

	w := performJSONRequest(router, http.MethodPost, "/admin/events/"+eventID.String()+"/publish", nil)

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "event not found")
	svc.AssertExpectations(t)
}

func TestAdminEventHandlerPublishEventInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.POST("/admin/events/:id/publish", h.PublishEvent)

	eventID := uuid.New()
	svc.On("PublishEvent", mock.Anything, eventID).Return((*model.Event)(nil), errors.New("db error")).Once()

	w := performJSONRequest(router, http.MethodPost, "/admin/events/"+eventID.String()+"/publish", nil)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to publish event")
	svc.AssertExpectations(t)
}

func TestAdminEventHandlerGetEventsStatsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.GET("/admin/events/stats", h.GetEventsStats)

	svc.On("GetEventsStats", mock.Anything).
		Return(map[string]int64{"total": 15, "published": 10}, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/events/stats", nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"total"`)
	svc.AssertExpectations(t)
}

func TestAdminEventHandlerGetEventsStatsInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(adminEventServiceMock)
	h := NewAdminEventHandler(svc)
	router := gin.New()
	router.GET("/admin/events/stats", h.GetEventsStats)

	svc.On("GetEventsStats", mock.Anything).
		Return((map[string]int64)(nil), errors.New("db error")).Once()

	w := performJSONRequest(router, http.MethodGet, "/admin/events/stats", nil)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to get stats")
	svc.AssertExpectations(t)
}
