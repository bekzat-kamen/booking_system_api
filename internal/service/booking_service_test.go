package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type bookingRepositoryMock struct {
	mock.Mock
}

func (m *bookingRepositoryMock) Create(ctx context.Context, booking *model.Booking, seats []*model.BookingSeat) error {
	args := m.Called(ctx, booking, seats)
	return args.Error(0)
}

func (m *bookingRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	args := m.Called(ctx, id)
	booking, _ := args.Get(0).(*model.Booking)
	return booking, args.Error(1)
}

func (m *bookingRepositoryMock) GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Booking, error) {
	args := m.Called(ctx, userID, limit, offset)
	bookings, _ := args.Get(0).([]*model.Booking)
	return bookings, args.Error(1)
}

func (m *bookingRepositoryMock) GetByEvent(ctx context.Context, eventID uuid.UUID, limit, offset int) ([]*model.Booking, error) {
	args := m.Called(ctx, eventID, limit, offset)
	bookings, _ := args.Get(0).([]*model.Booking)
	return bookings, args.Error(1)
}

func (m *bookingRepositoryMock) GetSeats(ctx context.Context, bookingID uuid.UUID) ([]*model.BookingSeat, error) {
	args := m.Called(ctx, bookingID)
	seats, _ := args.Get(0).([]*model.BookingSeat)
	return seats, args.Error(1)
}

func (m *bookingRepositoryMock) Update(ctx context.Context, booking *model.Booking) error {
	args := m.Called(ctx, booking)
	return args.Error(0)
}

func (m *bookingRepositoryMock) UpdateStatus(ctx context.Context, id uuid.UUID, status model.BookingStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *bookingRepositoryMock) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *bookingRepositoryMock) GetExpiredPending(ctx context.Context) ([]*model.Booking, error) {
	args := m.Called(ctx)
	bookings, _ := args.Get(0).([]*model.Booking)
	return bookings, args.Error(1)
}

type seatRepositoryMock struct {
	mock.Mock
}

func (m *seatRepositoryMock) Create(ctx context.Context, seat *model.Seat) error {
	args := m.Called(ctx, seat)
	return args.Error(0)
}

func (m *seatRepositoryMock) CreateBatch(ctx context.Context, seats []*model.Seat) error {
	args := m.Called(ctx, seats)
	return args.Error(0)
}

func (m *seatRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*model.Seat, error) {
	args := m.Called(ctx, id)
	seat, _ := args.Get(0).(*model.Seat)
	return seat, args.Error(1)
}

func (m *seatRepositoryMock) GetByEvent(ctx context.Context, eventID uuid.UUID) ([]*model.Seat, error) {
	args := m.Called(ctx, eventID)
	seats, _ := args.Get(0).([]*model.Seat)
	return seats, args.Error(1)
}

func (m *seatRepositoryMock) GetByEventAndStatus(ctx context.Context, eventID uuid.UUID, status model.SeatStatus) ([]*model.Seat, error) {
	args := m.Called(ctx, eventID, status)
	seats, _ := args.Get(0).([]*model.Seat)
	return seats, args.Error(1)
}

func (m *seatRepositoryMock) Update(ctx context.Context, seat *model.Seat) error {
	args := m.Called(ctx, seat)
	return args.Error(0)
}

func (m *seatRepositoryMock) UpdateStatus(ctx context.Context, id uuid.UUID, status model.SeatStatus, version int) error {
	args := m.Called(ctx, id, status, version)
	return args.Error(0)
}

func (m *seatRepositoryMock) CountByEvent(ctx context.Context, eventID uuid.UUID) (int, error) {
	args := m.Called(ctx, eventID)
	return args.Int(0), args.Error(1)
}

func (m *seatRepositoryMock) CountByStatus(ctx context.Context, eventID uuid.UUID, status model.SeatStatus) (int, error) {
	args := m.Called(ctx, eventID, status)
	return args.Int(0), args.Error(1)
}

func (m *seatRepositoryMock) Exists(ctx context.Context, eventID uuid.UUID, rowNumber, seatNumber string) (bool, error) {
	args := m.Called(ctx, eventID, rowNumber, seatNumber)
	return args.Bool(0), args.Error(1)
}

type eventRepositoryMock struct {
	mock.Mock
}

func (m *eventRepositoryMock) Create(ctx context.Context, event *model.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *eventRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*model.Event, error) {
	args := m.Called(ctx, id)
	event, _ := args.Get(0).(*model.Event)
	return event, args.Error(1)
}

func (m *eventRepositoryMock) GetAll(ctx context.Context, limit, offset int) ([]*model.Event, error) {
	args := m.Called(ctx, limit, offset)
	events, _ := args.Get(0).([]*model.Event)
	return events, args.Error(1)
}

func (m *eventRepositoryMock) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *eventRepositoryMock) GetByOrganizer(ctx context.Context, organizerID uuid.UUID, limit, offset int) ([]*model.Event, error) {
	args := m.Called(ctx, organizerID, limit, offset)
	events, _ := args.Get(0).([]*model.Event)
	return events, args.Error(1)
}

func (m *eventRepositoryMock) Update(ctx context.Context, event *model.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *eventRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestBookingServiceCreateBookingSuccess(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	eventID := uuid.New()
	seatID1 := uuid.New()
	seatID2 := uuid.New()

	bookingRepo := new(bookingRepositoryMock)
	seatRepo := new(seatRepositoryMock)
	eventRepo := new(eventRepositoryMock)
	svc := NewBookingService(bookingRepo, seatRepo, eventRepo)

	event := &model.Event{
		ID:             eventID,
		Title:          "Concert",
		AvailableSeats: 10,
	}
	seat1 := &model.Seat{ID: seatID1, EventID: eventID, SeatNumber: "1", RowNumber: "A", Price: 100, Status: model.SeatStatusAvailable}
	seat2 := &model.Seat{ID: seatID2, EventID: eventID, SeatNumber: "2", RowNumber: "A", Price: 150, Status: model.SeatStatusAvailable}

	req := &model.CreateBookingRequest{
		EventID: eventID,
		SeatIDs: []uuid.UUID{seatID1, seatID2},
	}

	eventRepo.On("GetByID", ctx, eventID).Return(event, nil).Twice()
	seatRepo.On("GetByID", ctx, seatID1).Return(seat1, nil).Twice()
	seatRepo.On("GetByID", ctx, seatID2).Return(seat2, nil).Twice()

	bookingRepo.On("Create", ctx, mock.MatchedBy(func(booking *model.Booking) bool {
		return booking.UserID == userID &&
			booking.EventID == eventID &&
			booking.Status == model.BookingStatusPending &&
			booking.TotalAmount == 250 &&
			booking.FinalAmount == 250
	}), mock.MatchedBy(func(seats []*model.BookingSeat) bool {
		return len(seats) == 2 && seats[0].SeatID == seatID1 && seats[1].SeatID == seatID2
	})).Return(nil).Once()

	seatRepo.On("Update", ctx, seat1).Return(nil).Once()
	seatRepo.On("Update", ctx, seat2).Return(nil).Once()
	eventRepo.On("Update", ctx, mock.MatchedBy(func(updated *model.Event) bool {
		return updated.ID == eventID && updated.AvailableSeats == 8
	})).Return(nil).Once()

	bookingRepo.On("GetByID", ctx, mock.AnythingOfType("uuid.UUID")).Return(&model.Booking{
		ID:          uuid.New(),
		UserID:      userID,
		EventID:     eventID,
		TotalAmount: 250,
		Discount:    0,
		FinalAmount: 250,
		Status:      model.BookingStatusPending,
		ExpiresAt:   time.Now().Add(15 * time.Minute),
		CreatedAt:   time.Now(),
	}, nil).Once()
	bookingRepo.On("GetSeats", ctx, mock.AnythingOfType("uuid.UUID")).Return([]*model.BookingSeat{
		{SeatID: seatID1, Price: 100},
		{SeatID: seatID2, Price: 150},
	}, nil).Once()

	resp, err := svc.CreateBooking(ctx, userID, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "Concert", resp.EventTitle)
	assert.Len(t, resp.Seats, 2)
	assert.Equal(t, 250.0, resp.FinalAmount)
	assert.Equal(t, model.BookingStatusPending, resp.Status)

	bookingRepo.AssertExpectations(t)
	seatRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestBookingServiceCreateBookingRollbackOnSeatUpdateFailure(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	eventID := uuid.New()
	seatID := uuid.New()

	bookingRepo := new(bookingRepositoryMock)
	seatRepo := new(seatRepositoryMock)
	eventRepo := new(eventRepositoryMock)
	svc := NewBookingService(bookingRepo, seatRepo, eventRepo)

	event := &model.Event{ID: eventID, AvailableSeats: 5}
	seat := &model.Seat{ID: seatID, EventID: eventID, Price: 100, Status: model.SeatStatusAvailable}
	req := &model.CreateBookingRequest{EventID: eventID, SeatIDs: []uuid.UUID{seatID}}

	eventRepo.On("GetByID", ctx, eventID).Return(event, nil).Once()
	seatRepo.On("GetByID", ctx, seatID).Return(seat, nil).Once()
	bookingRepo.On("Create", ctx, mock.AnythingOfType("*model.Booking"), mock.AnythingOfType("[]*model.BookingSeat")).Return(nil).Once()
	seatRepo.On("Update", ctx, seat).Return(errors.New("update failed")).Once()
	bookingRepo.On("UpdateStatus", ctx, mock.AnythingOfType("uuid.UUID"), model.BookingStatusCancelled).Return(nil).Once()

	resp, err := svc.CreateBooking(ctx, userID, req)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "failed to reserve seats")
	eventRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)

	bookingRepo.AssertExpectations(t)
	seatRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestBookingServiceConfirmBookingExpired(t *testing.T) {
	ctx := context.Background()
	bookingID := uuid.New()

	bookingRepo := new(bookingRepositoryMock)
	seatRepo := new(seatRepositoryMock)
	eventRepo := new(eventRepositoryMock)
	svc := NewBookingService(bookingRepo, seatRepo, eventRepo)

	bookingRepo.On("GetByID", ctx, bookingID).Return(&model.Booking{
		ID:        bookingID,
		Status:    model.BookingStatusPending,
		ExpiresAt: time.Now().Add(-time.Minute),
	}, nil).Once()

	err := svc.ConfirmBooking(ctx, bookingID)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrBookingExpired)
	bookingRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	bookingRepo.AssertExpectations(t)
}

func TestBookingServiceGetBooking(t *testing.T) {
	ctx := context.Background()
	bookingID := uuid.New()

	bookingRepo := new(bookingRepositoryMock)
	seatRepo := new(seatRepositoryMock)
	eventRepo := new(eventRepositoryMock)
	svc := NewBookingService(bookingRepo, seatRepo, eventRepo)

	t.Run("Success", func(t *testing.T) {
		bookingRepo.On("GetByID", ctx, bookingID).Return(&model.Booking{ID: bookingID, EventID: uuid.New()}, nil).Twice()
		eventRepo.On("GetByID", ctx, mock.Anything).Return(&model.Event{Title: "Concert"}, nil).Once()
		bookingRepo.On("GetSeats", ctx, bookingID).Return([]*model.BookingSeat{}, nil).Once()

		resp, err := svc.GetBooking(ctx, bookingID)
		assert.NoError(t, err)
		assert.Equal(t, bookingID, resp.ID)
	})
}

func TestBookingServiceGetUserBookings(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	bookingID := uuid.New()

	bookingRepo := new(bookingRepositoryMock)
	seatRepo := new(seatRepositoryMock)
	eventRepo := new(eventRepositoryMock)
	svc := NewBookingService(bookingRepo, seatRepo, eventRepo)

	t.Run("Success", func(t *testing.T) {
		bookingRepo.On("GetByUser", ctx, userID, 10, 0).Return([]*model.Booking{{ID: bookingID, EventID: uuid.New()}}, nil).Once()
		bookingRepo.On("CountByUser", ctx, userID).Return(1, nil).Once()
		bookingRepo.On("GetByID", ctx, bookingID).Return(&model.Booking{ID: bookingID, EventID: uuid.New()}, nil).Once()
		bookingRepo.On("GetSeats", ctx, bookingID).Return([]*model.BookingSeat{}, nil).Once()
		eventRepo.On("GetByID", ctx, mock.Anything).Return(&model.Event{Title: "Concert"}, nil).Once()

		resp, total, err := svc.GetUserBookings(ctx, userID, 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, resp, 1)
	})
}

func TestBookingServiceCancelBooking(t *testing.T) {
	ctx := context.Background()
	bookingID := uuid.New()
	userID := uuid.New()
	eventID := uuid.New()

	bookingRepo := new(bookingRepositoryMock)
	seatRepo := new(seatRepositoryMock)
	eventRepo := new(eventRepositoryMock)
	svc := NewBookingService(bookingRepo, seatRepo, eventRepo)

	t.Run("Success", func(t *testing.T) {
		bookingRepo.On("GetByID", ctx, bookingID).Return(&model.Booking{
			ID:      bookingID,
			UserID:  userID,
			Status:  model.BookingStatusPending,
			EventID: eventID,
		}, nil).Once()
		bookingRepo.On("GetSeats", ctx, bookingID).Return([]*model.BookingSeat{{SeatID: uuid.New()}}, nil).Once()
		bookingRepo.On("Update", ctx, mock.MatchedBy(func(b *model.Booking) bool {
			return b.Status == model.BookingStatusCancelled
		})).Return(nil).Once()
		seatRepo.On("GetByID", ctx, mock.Anything).Return(&model.Seat{Status: model.SeatStatusReserved}, nil).Once()
		seatRepo.On("Update", ctx, mock.Anything).Return(nil).Once()
		eventRepo.On("GetByID", ctx, eventID).Return(&model.Event{AvailableSeats: 5}, nil).Once()
		eventRepo.On("Update", ctx, mock.Anything).Return(nil).Once()

		err := svc.CancelBooking(ctx, bookingID, userID)
		assert.NoError(t, err)
	})
}

func TestBookingServiceExpirePendingBookings(t *testing.T) {
	ctx := context.Background()

	bookingRepo := new(bookingRepositoryMock)
	seatRepo := new(seatRepositoryMock)
	eventRepo := new(eventRepositoryMock)
	svc := NewBookingService(bookingRepo, seatRepo, eventRepo)

	t.Run("Success", func(t *testing.T) {
		bookingRepo.On("GetExpiredPending", ctx).Return([]*model.Booking{{ID: uuid.New(), UserID: uuid.New(), EventID: uuid.New()}}, nil).Once()
		bookingRepo.On("GetSeats", ctx, mock.Anything).Return([]*model.BookingSeat{{SeatID: uuid.New()}}, nil).Once()
		bookingRepo.On("Update", ctx, mock.MatchedBy(func(b *model.Booking) bool {
			return b.Status == model.BookingStatusExpired
		})).Return(nil).Once()
		seatRepo.On("GetByID", ctx, mock.Anything).Return(&model.Seat{Status: model.SeatStatusReserved}, nil).Once()
		seatRepo.On("Update", ctx, mock.Anything).Return(nil).Once()
		eventRepo.On("GetByID", ctx, mock.Anything).Return(&model.Event{AvailableSeats: 5}, nil).Once()
		eventRepo.On("Update", ctx, mock.Anything).Return(nil).Once()

		count, err := svc.ExpirePendingBookings(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

