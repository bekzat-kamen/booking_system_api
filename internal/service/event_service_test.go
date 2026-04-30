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

func TestEventServiceCreateInvalidDate(t *testing.T) {
	svc := NewEventService(new(eventRepositoryMock))

	event, err := svc.Create(context.Background(), uuid.New(), &model.CreateEventRequest{
		Title:      "Event",
		Venue:      "Venue",
		EventDate:  "bad-date",
		TotalSeats: 10,
		Price:      100,
	})

	require.Error(t, err)
	assert.Nil(t, event)
	assert.ErrorIs(t, err, ErrInvalidEventDate)
}

func TestEventServiceCreateSuccess(t *testing.T) {
	ctx := context.Background()
	repo := new(eventRepositoryMock)
	svc := NewEventService(repo)
	organizerID := uuid.New()

	repo.On("Create", ctx, mock.MatchedBy(func(event *model.Event) bool {
		return event.Title == "Concert" &&
			event.CreatedBy == organizerID &&
			event.TotalSeats == 20 &&
			event.AvailableSeats == 20 &&
			event.Status == model.EventStatusDraft
	})).Return(nil).Once()
	repo.On("GetByID", ctx, mock.AnythingOfType("uuid.UUID")).Return(&model.Event{
		ID:             uuid.New(),
		Title:          "Concert",
		CreatedBy:      organizerID,
		TotalSeats:     20,
		AvailableSeats: 20,
		Status:         model.EventStatusDraft,
	}, nil).Once()

	event, err := svc.Create(ctx, organizerID, &model.CreateEventRequest{
		Title:      "Concert",
		Venue:      "Hall",
		EventDate:  time.Now().Add(time.Hour).Format(time.RFC3339),
		TotalSeats: 20,
		Price:      100,
	})

	require.NoError(t, err)
	require.NotNil(t, event)
	assert.Equal(t, "Concert", event.Title)
	repo.AssertExpectations(t)
}

func TestEventServiceUpdateForbidden(t *testing.T) {
	ctx := context.Background()
	repo := new(eventRepositoryMock)
	svc := NewEventService(repo)
	eventID := uuid.New()
	ownerID := uuid.New()
	otherID := uuid.New()

	repo.On("GetByID", ctx, eventID).Return(&model.Event{ID: eventID, CreatedBy: ownerID}, nil).Once()

	event, err := svc.Update(ctx, eventID, otherID, &model.UpdateEventRequest{Title: "New"})

	require.Error(t, err)
	assert.Nil(t, event)
	assert.ErrorIs(t, err, ErrEventForbidden)
	repo.AssertExpectations(t)
}

func TestEventServiceUpdateInvalidSeatBalance(t *testing.T) {
	ctx := context.Background()
	repo := new(eventRepositoryMock)
	svc := NewEventService(repo)
	eventID := uuid.New()
	ownerID := uuid.New()

	repo.On("GetByID", ctx, eventID).Return(&model.Event{
		ID:             eventID,
		CreatedBy:      ownerID,
		TotalSeats:     10,
		AvailableSeats: 2,
	}, nil).Once()

	event, err := svc.Update(ctx, eventID, ownerID, &model.UpdateEventRequest{TotalSeats: 5})

	require.Error(t, err)
	assert.Nil(t, event)
	assert.ErrorIs(t, err, ErrInvalidSeatBalance)
	repo.AssertExpectations(t)
}

func TestEventServicePublishEventSoldOut(t *testing.T) {
	ctx := context.Background()
	repo := new(eventRepositoryMock)
	svc := NewEventService(repo)
	eventID := uuid.New()
	ownerID := uuid.New()

	repo.On("GetByID", ctx, eventID).Return(&model.Event{
		ID:             eventID,
		CreatedBy:      ownerID,
		AvailableSeats: 0,
	}, nil).Once()
	repo.On("Update", ctx, mock.MatchedBy(func(event *model.Event) bool {
		return event.ID == eventID && event.Status == model.EventStatusSoldOut
	})).Return(nil).Once()
	repo.On("GetByID", ctx, eventID).Return(&model.Event{
		ID:        eventID,
		CreatedBy: ownerID,
		Status:    model.EventStatusSoldOut,
	}, nil).Once()

	event, err := svc.PublishEvent(ctx, eventID, ownerID)

	require.NoError(t, err)
	require.NotNil(t, event)
	assert.Equal(t, model.EventStatusSoldOut, event.Status)
	repo.AssertExpectations(t)
}
