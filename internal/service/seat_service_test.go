package service

import (
	"context"
	"testing"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSeatServiceGenerateSeatsAlreadyGenerated(t *testing.T) {
	ctx := context.Background()
	seatRepo := new(seatRepositoryMock)
	eventRepo := new(eventRepositoryMock)
	svc := NewSeatService(seatRepo, eventRepo)
	eventID := uuid.New()
	ownerID := uuid.New()

	eventRepo.On("GetByID", ctx, eventID).Return(&model.Event{ID: eventID, CreatedBy: ownerID}, nil).Once()
	seatRepo.On("GetByEvent", ctx, eventID).Return([]*model.Seat{{ID: uuid.New()}}, nil).Once()

	err := svc.GenerateSeats(ctx, eventID, ownerID, &model.GenerateSeatsRequest{TotalRows: 2, SeatsPerRow: 2, BasePrice: 10})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSeatsAlreadyGenerated)
}

func TestSeatServiceGenerateSeatsSuccess(t *testing.T) {
	ctx := context.Background()
	seatRepo := new(seatRepositoryMock)
	eventRepo := new(eventRepositoryMock)
	svc := NewSeatService(seatRepo, eventRepo)
	eventID := uuid.New()
	ownerID := uuid.New()

	event := &model.Event{ID: eventID, CreatedBy: ownerID}
	eventRepo.On("GetByID", ctx, eventID).Return(event, nil).Once()
	seatRepo.On("GetByEvent", ctx, eventID).Return([]*model.Seat{}, nil).Once()
	seatRepo.On("CreateBatch", ctx, mock.MatchedBy(func(seats []*model.Seat) bool {
		return len(seats) == 6 && seats[0].EventID == eventID && seats[0].Status == model.SeatStatusAvailable
	})).Return(nil).Once()
	eventRepo.On("Update", ctx, mock.MatchedBy(func(updated *model.Event) bool {
		return updated.TotalSeats == 6 && updated.AvailableSeats == 6
	})).Return(nil).Once()

	err := svc.GenerateSeats(ctx, eventID, ownerID, &model.GenerateSeatsRequest{TotalRows: 2, SeatsPerRow: 3, BasePrice: 10})

	require.NoError(t, err)
	seatRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestSeatServiceGetSeatMapSuccess(t *testing.T) {
	ctx := context.Background()
	seatRepo := new(seatRepositoryMock)
	eventRepo := new(eventRepositoryMock)
	svc := NewSeatService(seatRepo, eventRepo)
	eventID := uuid.New()

	eventRepo.On("GetByID", ctx, eventID).Return(&model.Event{ID: eventID}, nil).Once()
	seatRepo.On("GetByEvent", ctx, eventID).Return([]*model.Seat{
		{ID: uuid.New(), EventID: eventID, Status: model.SeatStatusAvailable},
		{ID: uuid.New(), EventID: eventID, Status: model.SeatStatusBooked},
		{ID: uuid.New(), EventID: eventID, Status: model.SeatStatusReserved},
		{ID: uuid.New(), EventID: eventID, Status: model.SeatStatusBlocked},
	}, nil).Once()

	resp, err := svc.GetSeatMap(ctx, eventID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 4, resp.TotalSeats)
	assert.Equal(t, 1, resp.Available)
	assert.Equal(t, 1, resp.Booked)
	assert.Equal(t, 1, resp.Reserved)
	assert.Equal(t, 1, resp.Blocked)
}

func TestSeatServiceReserveSeatNotAvailable(t *testing.T) {
	ctx := context.Background()
	seatRepo := new(seatRepositoryMock)
	eventRepo := new(eventRepositoryMock)
	svc := NewSeatService(seatRepo, eventRepo)
	seatID := uuid.New()

	seatRepo.On("GetByID", ctx, seatID).Return(&model.Seat{ID: seatID, Status: model.SeatStatusBooked}, nil).Once()

	err := svc.ReserveSeat(ctx, seatID)

	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrSeatNotAvailable)
}

func TestSeatServiceReleaseSeatReserved(t *testing.T) {
	ctx := context.Background()
	seatRepo := new(seatRepositoryMock)
	eventRepo := new(eventRepositoryMock)
	svc := NewSeatService(seatRepo, eventRepo)
	seatID := uuid.New()
	seat := &model.Seat{ID: seatID, Status: model.SeatStatusReserved}

	seatRepo.On("GetByID", ctx, seatID).Return(seat, nil).Once()
	seatRepo.On("Update", ctx, seat).Return(nil).Once()

	err := svc.ReleaseSeat(ctx, seatID)

	require.NoError(t, err)
	assert.Equal(t, model.SeatStatusAvailable, seat.Status)
	seatRepo.AssertExpectations(t)
}

func TestSeatServiceGetAvailableSeatsSuccess(t *testing.T) {
	ctx := context.Background()
	seatRepo := new(seatRepositoryMock)
	svc := NewSeatService(seatRepo, new(eventRepositoryMock))
	eventID := uuid.New()

	seatRepo.On("GetByEventAndStatus", ctx, eventID, model.SeatStatusAvailable).Return([]*model.Seat{
		{ID: uuid.New(), Status: model.SeatStatusAvailable},
	}, nil).Once()

	seats, err := svc.GetAvailableSeats(ctx, eventID)
	assert.NoError(t, err)
	assert.Len(t, seats, 1)
}

func TestSeatServiceBookSeatSuccess(t *testing.T) {
	ctx := context.Background()
	seatRepo := new(seatRepositoryMock)
	svc := NewSeatService(seatRepo, new(eventRepositoryMock))
	seatID := uuid.New()
	seat := &model.Seat{ID: seatID, Status: model.SeatStatusReserved}

	seatRepo.On("GetByID", ctx, seatID).Return(seat, nil).Once()
	seatRepo.On("Update", ctx, seat).Return(nil).Once()

	err := svc.BookSeat(ctx, seatID)
	assert.NoError(t, err)
	assert.Equal(t, model.SeatStatusBooked, seat.Status)
}
