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

type adminEventRepositoryMock struct {
	mock.Mock
}

func (m *adminEventRepositoryMock) GetAllEvents(ctx context.Context, page, limit int, status, organizerID string) ([]*model.Event, int, error) {
	args := m.Called(ctx, page, limit, status, organizerID)
	resp, _ := args.Get(0).([]*model.Event)
	return resp, args.Int(1), args.Error(2)
}

func (m *adminEventRepositoryMock) GetEventByID(ctx context.Context, id uuid.UUID) (*model.Event, error) {
	args := m.Called(ctx, id)
	resp, _ := args.Get(0).(*model.Event)
	return resp, args.Error(1)
}

func (m *adminEventRepositoryMock) UpdateEventAdmin(ctx context.Context, event *model.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *adminEventRepositoryMock) DeleteEventAdmin(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *adminEventRepositoryMock) PublishEventAdmin(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *adminEventRepositoryMock) GetEventStats(ctx context.Context, eventID uuid.UUID) (map[string]interface{}, error) {
	args := m.Called(ctx, eventID)
	resp, _ := args.Get(0).(map[string]interface{})
	return resp, args.Error(1)
}

func (m *adminEventRepositoryMock) GetEventsByStatus(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	resp, _ := args.Get(0).(map[string]int64)
	return resp, args.Error(1)
}

func TestAdminEventServiceGetEventDetailSuccess(t *testing.T) {
	ctx := context.Background()
	repo := new(adminEventRepositoryMock)
	svc := NewAdminEventService(repo)
	eventID := uuid.New()

	repo.On("GetEventByID", ctx, eventID).Return(&model.Event{ID: eventID, Title: "Concert"}, nil).Once()
	repo.On("GetEventStats", ctx, eventID).Return(map[string]interface{}{"total_bookings": int64(10)}, nil).Once()

	detail, err := svc.GetEventDetail(ctx, eventID)

	require.NoError(t, err)
	assert.NotNil(t, detail["event"])
	assert.NotNil(t, detail["statistics"])
}

func TestAdminEventServiceUpdateEventInvalidDate(t *testing.T) {
	ctx := context.Background()
	repo := new(adminEventRepositoryMock)
	svc := NewAdminEventService(repo)
	eventID := uuid.New()

	repo.On("GetEventByID", ctx, eventID).Return(&model.Event{ID: eventID}, nil).Once()

	event, err := svc.UpdateEvent(ctx, eventID, &model.UpdateEventRequest{EventDate: "bad-date"})

	require.Error(t, err)
	assert.Nil(t, event)
	assert.EqualError(t, err, "invalid event date format")
}

func TestAdminEventServiceUpdateEventSuccess(t *testing.T) {
	ctx := context.Background()
	repo := new(adminEventRepositoryMock)
	svc := NewAdminEventService(repo)
	eventID := uuid.New()
	event := &model.Event{ID: eventID, AvailableSeats: 15, TotalSeats: 20, Title: "Old"}

	repo.On("GetEventByID", ctx, eventID).Return(event, nil).Once()
	repo.On("UpdateEventAdmin", ctx, mock.MatchedBy(func(updated *model.Event) bool {
		return updated.Title == "New" && updated.TotalSeats == 10 && updated.AvailableSeats == 10
	})).Return(nil).Once()

	resp, err := svc.UpdateEvent(ctx, eventID, &model.UpdateEventRequest{
		Title:      "New",
		TotalSeats: 10,
		EventDate:  time.Now().Add(time.Hour).Format(time.RFC3339),
	})

	require.NoError(t, err)
	assert.Equal(t, "New", resp.Title)
}

func TestAdminEventServicePublishEventSuccess(t *testing.T) {
	ctx := context.Background()
	repo := new(adminEventRepositoryMock)
	svc := NewAdminEventService(repo)
	eventID := uuid.New()

	repo.On("PublishEventAdmin", ctx, eventID).Return(nil).Once()
	repo.On("GetEventByID", ctx, eventID).Return(&model.Event{ID: eventID, Status: model.EventStatusPublished}, nil).Once()

	event, err := svc.PublishEvent(ctx, eventID)

	require.NoError(t, err)
	assert.Equal(t, model.EventStatusPublished, event.Status)
}
