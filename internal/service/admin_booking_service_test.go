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

type adminBookingRepositoryMock struct {
	mock.Mock
}

func (m *adminBookingRepositoryMock) GetAllBookings(ctx context.Context, page, limit int, status, userID, eventID string) ([]*model.Booking, int, error) {
	args := m.Called(ctx, page, limit, status, userID, eventID)
	resp, _ := args.Get(0).([]*model.Booking)
	return resp, args.Int(1), args.Error(2)
}

func (m *adminBookingRepositoryMock) GetBookingByID(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	args := m.Called(ctx, id)
	resp, _ := args.Get(0).(*model.Booking)
	return resp, args.Error(1)
}

func (m *adminBookingRepositoryMock) GetBookingSeats(ctx context.Context, bookingID uuid.UUID) ([]*model.BookingSeat, error) {
	args := m.Called(ctx, bookingID)
	resp, _ := args.Get(0).([]*model.BookingSeat)
	return resp, args.Error(1)
}

func (m *adminBookingRepositoryMock) GetBookingPayment(ctx context.Context, bookingID uuid.UUID) (*model.Payment, error) {
	args := m.Called(ctx, bookingID)
	resp, _ := args.Get(0).(*model.Payment)
	return resp, args.Error(1)
}

func (m *adminBookingRepositoryMock) CancelBookingAdmin(ctx context.Context, id uuid.UUID, status model.BookingStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *adminBookingRepositoryMock) RefundPayment(ctx context.Context, paymentID uuid.UUID) error {
	args := m.Called(ctx, paymentID)
	return args.Error(0)
}

func (m *adminBookingRepositoryMock) GetBookingsStats(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	resp, _ := args.Get(0).(map[string]int64)
	return resp, args.Error(1)
}

func (m *adminBookingRepositoryMock) GetRevenueStats(ctx context.Context) (map[string]float64, error) {
	args := m.Called(ctx)
	resp, _ := args.Get(0).(map[string]float64)
	return resp, args.Error(1)
}

func TestAdminBookingServiceGetBookingDetailSuccess(t *testing.T) {
	ctx := context.Background()
	repo := new(adminBookingRepositoryMock)
	svc := NewAdminBookingService(repo, new(seatRepositoryMock), new(eventRepositoryMock))
	id := uuid.New()

	repo.On("GetBookingByID", ctx, id).Return(&model.Booking{ID: id, Email: "user@example.com"}, nil).Once()
	repo.On("GetBookingSeats", ctx, id).Return([]*model.BookingSeat{{SeatNumber: "1", RowNumber: "A"}}, nil).Once()
	repo.On("GetBookingPayment", ctx, id).Return(&model.Payment{ID: uuid.New()}, nil).Once()

	detail, err := svc.GetBookingDetail(ctx, id)

	require.NoError(t, err)
	assert.NotNil(t, detail["booking"])
	assert.NotNil(t, detail["seats"])
	assert.NotNil(t, detail["payment"])
}

func TestAdminBookingServiceCancelBookingCannotCancel(t *testing.T) {
	ctx := context.Background()
	repo := new(adminBookingRepositoryMock)
	svc := NewAdminBookingService(repo, new(seatRepositoryMock), new(eventRepositoryMock))
	id := uuid.New()

	repo.On("GetBookingByID", ctx, id).Return(&model.Booking{ID: id, Status: model.BookingStatusCompleted}, nil).Once()

	err := svc.CancelBooking(ctx, id, false)

	require.Error(t, err)
	assert.EqualError(t, err, "cannot cancel this booking")
}

func TestAdminBookingServiceRefundBookingAlreadyRefunded(t *testing.T) {
	ctx := context.Background()
	repo := new(adminBookingRepositoryMock)
	svc := NewAdminBookingService(repo, new(seatRepositoryMock), new(eventRepositoryMock))
	id := uuid.New()

	repo.On("GetBookingPayment", ctx, id).Return(&model.Payment{ID: uuid.New(), Status: model.PaymentStatusRefunded}, nil).Once()

	err := svc.RefundBooking(ctx, id)

	require.Error(t, err)
	assert.EqualError(t, err, "payment already refunded")
}

func TestAdminBookingServiceExportBookingsToCSVSuccess(t *testing.T) {
	ctx := context.Background()
	repo := new(adminBookingRepositoryMock)
	svc := NewAdminBookingService(repo, new(seatRepositoryMock), new(eventRepositoryMock))
	bookingID := uuid.New()
	now := time.Now()

	repo.On("GetAllBookings", ctx, 1, 10000, "confirmed", "", "").Return([]*model.Booking{
		{
			ID:          bookingID,
			Email:       "user@example.com",
			EventTitle:  "Concert",
			EventDate:   now,
			TotalAmount: 100,
			Discount:    10,
			FinalAmount: 90,
			Status:      model.BookingStatusConfirmed,
			CreatedAt:   now,
		},
	}, 1, nil).Once()
	repo.On("GetBookingSeats", ctx, bookingID).Return([]*model.BookingSeat{
		{RowNumber: "A", SeatNumber: "1"},
		{RowNumber: "A", SeatNumber: "2"},
	}, nil).Once()

	rows, err := svc.ExportBookingsToCSV(ctx, "confirmed")

	require.NoError(t, err)
	assert.Len(t, rows, 2)
	assert.Equal(t, "Booking ID", rows[0][0])
	assert.Equal(t, bookingID.String(), rows[1][0])
}
