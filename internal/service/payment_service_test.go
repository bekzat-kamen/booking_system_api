package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type paymentRepositoryMock struct {
	mock.Mock
}

func (m *paymentRepositoryMock) Create(ctx context.Context, payment *model.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *paymentRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*model.Payment, error) {
	args := m.Called(ctx, id)
	payment, _ := args.Get(0).(*model.Payment)
	return payment, args.Error(1)
}

func (m *paymentRepositoryMock) GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*model.Payment, error) {
	args := m.Called(ctx, bookingID)
	payment, _ := args.Get(0).(*model.Payment)
	return payment, args.Error(1)
}

func (m *paymentRepositoryMock) GetByTransactionID(ctx context.Context, transactionID string) (*model.Payment, error) {
	args := m.Called(ctx, transactionID)
	payment, _ := args.Get(0).(*model.Payment)
	return payment, args.Error(1)
}

func (m *paymentRepositoryMock) Update(ctx context.Context, payment *model.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *paymentRepositoryMock) UpdateStatus(ctx context.Context, id uuid.UUID, status model.PaymentStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *paymentRepositoryMock) SetProviderResponse(ctx context.Context, id uuid.UUID, response interface{}) error {
	args := m.Called(ctx, id, response)
	return args.Error(0)
}

type bookingServiceInterfaceMock struct {
	mock.Mock
}

func (m *bookingServiceInterfaceMock) CreateBooking(ctx context.Context, userID uuid.UUID, req *model.CreateBookingRequest) (*model.BookingResponse, error) {
	args := m.Called(ctx, userID, req)
	resp, _ := args.Get(0).(*model.BookingResponse)
	return resp, args.Error(1)
}

func (m *bookingServiceInterfaceMock) GetBooking(ctx context.Context, id uuid.UUID) (*model.BookingResponse, error) {
	args := m.Called(ctx, id)
	resp, _ := args.Get(0).(*model.BookingResponse)
	return resp, args.Error(1)
}

func (m *bookingServiceInterfaceMock) GetUserBookings(ctx context.Context, userID uuid.UUID, page, limit int) ([]*model.BookingResponse, int, error) {
	args := m.Called(ctx, userID, page, limit)
	resp, _ := args.Get(0).([]*model.BookingResponse)
	return resp, args.Int(1), args.Error(2)
}

func (m *bookingServiceInterfaceMock) CancelBooking(ctx context.Context, bookingID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, bookingID, userID)
	return args.Error(0)
}

func (m *bookingServiceInterfaceMock) ConfirmBooking(ctx context.Context, bookingID uuid.UUID) error {
	args := m.Called(ctx, bookingID)
	return args.Error(0)
}

func (m *bookingServiceInterfaceMock) ExpirePendingBookings(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func TestPaymentServiceCreatePaymentSuccess(t *testing.T) {
	ctx := context.Background()
	bookingID := uuid.New()

	paymentRepo := new(paymentRepositoryMock)
	bookingRepo := new(bookingRepositoryMock)
	bookingSvc := new(bookingServiceInterfaceMock)
	svc := NewPaymentService(paymentRepo, bookingRepo, bookingSvc)

	req := &model.CreatePaymentRequest{
		BookingID:     bookingID,
		PaymentMethod: model.PaymentMethodCard,
	}

	bookingRepo.On("GetByID", ctx, bookingID).Return(&model.Booking{
		ID:          bookingID,
		FinalAmount: 499.99,
		Status:      model.BookingStatusPending,
	}, nil).Once()
	paymentRepo.On("GetByBookingID", ctx, bookingID).Return((*model.Payment)(nil), repository.ErrPaymentNotFound).Once()
	paymentRepo.On("Create", ctx, mock.MatchedBy(func(payment *model.Payment) bool {
		return payment.BookingID == bookingID &&
			payment.Amount == 499.99 &&
			payment.Status == model.PaymentStatusPending &&
			payment.PaymentMethod == model.PaymentMethodCard &&
			payment.Provider == "mock" &&
			payment.TransactionID != ""
	})).Return(nil).Once()

	payment, err := svc.CreatePayment(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, payment)
	assert.Equal(t, bookingID, payment.BookingID)
	assert.Equal(t, 499.99, payment.Amount)
	assert.Equal(t, model.PaymentStatusPending, payment.Status)

	paymentRepo.AssertExpectations(t)
	bookingRepo.AssertExpectations(t)
	bookingSvc.AssertExpectations(t)
}

func TestPaymentServiceCreatePaymentAlreadyExists(t *testing.T) {
	ctx := context.Background()
	bookingID := uuid.New()

	paymentRepo := new(paymentRepositoryMock)
	bookingRepo := new(bookingRepositoryMock)
	bookingSvc := new(bookingServiceInterfaceMock)
	svc := NewPaymentService(paymentRepo, bookingRepo, bookingSvc)

	bookingRepo.On("GetByID", ctx, bookingID).Return(&model.Booking{
		ID:          bookingID,
		FinalAmount: 100,
		Status:      model.BookingStatusPending,
	}, nil).Once()
	paymentRepo.On("GetByBookingID", ctx, bookingID).Return(&model.Payment{ID: uuid.New(), BookingID: bookingID}, nil).Once()

	payment, err := svc.CreatePayment(ctx, &model.CreatePaymentRequest{
		BookingID:     bookingID,
		PaymentMethod: model.PaymentMethodCard,
	})

	require.Error(t, err)
	assert.Nil(t, payment)
	assert.ErrorIs(t, err, ErrPaymentAlreadyExists)
	paymentRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)

	paymentRepo.AssertExpectations(t)
	bookingRepo.AssertExpectations(t)
}

func TestPaymentServiceProcessPaymentSuccess(t *testing.T) {
	ctx := context.Background()
	paymentID := uuid.New()
	bookingID := uuid.New()

	paymentRepo := new(paymentRepositoryMock)
	bookingRepo := new(bookingRepositoryMock)
	bookingSvc := new(bookingServiceInterfaceMock)
	svc := NewPaymentService(paymentRepo, bookingRepo, bookingSvc)

	payment := &model.Payment{
		ID:            paymentID,
		BookingID:     bookingID,
		Status:        model.PaymentStatusPending,
		PaymentMethod: model.PaymentMethodCard,
	}

	paymentRepo.On("GetByID", ctx, paymentID).Return(payment, nil).Once()
	paymentRepo.On("Update", ctx, mock.MatchedBy(func(updated *model.Payment) bool {
		return updated.ID == paymentID &&
			updated.Status == model.PaymentStatusSuccess &&
			updated.PaidAt != nil &&
			updated.ProviderResponse != ""
	})).Return(nil).Once()
	bookingSvc.On("ConfirmBooking", ctx, bookingID).Return(nil).Once()

	result, err := svc.ProcessPayment(ctx, paymentID)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, model.PaymentStatusSuccess, result.Status)
	assert.NotNil(t, result.PaidAt)

	paymentRepo.AssertExpectations(t)
	bookingSvc.AssertExpectations(t)
	bookingRepo.AssertExpectations(t)
}

func TestPaymentServiceProcessPaymentConfirmFailureMarksPaymentFailed(t *testing.T) {
	ctx := context.Background()
	paymentID := uuid.New()
	bookingID := uuid.New()

	paymentRepo := new(paymentRepositoryMock)
	bookingRepo := new(bookingRepositoryMock)
	bookingSvc := new(bookingServiceInterfaceMock)
	svc := NewPaymentService(paymentRepo, bookingRepo, bookingSvc)

	payment := &model.Payment{
		ID:        paymentID,
		BookingID: bookingID,
		Status:    model.PaymentStatusPending,
	}

	paymentRepo.On("GetByID", ctx, paymentID).Return(payment, nil).Once()
	paymentRepo.On("Update", ctx, mock.MatchedBy(func(updated *model.Payment) bool {
		return updated.ID == paymentID && updated.Status == model.PaymentStatusSuccess
	})).Return(nil).Once()
	bookingSvc.On("ConfirmBooking", ctx, bookingID).Return(errors.New("confirm failed")).Once()
	paymentRepo.On("Update", ctx, mock.MatchedBy(func(updated *model.Payment) bool {
		return updated.ID == paymentID && updated.Status == model.PaymentStatusFailed
	})).Return(nil).Once()

	result, err := svc.ProcessPayment(ctx, paymentID)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "failed to confirm booking")

	paymentRepo.AssertExpectations(t)
	bookingSvc.AssertExpectations(t)
	bookingRepo.AssertExpectations(t)
}

func TestPaymentServiceProcessWebhookSuccessConfirmsBooking(t *testing.T) {
	ctx := context.Background()
	bookingID := uuid.New()
	payment := &model.Payment{
		ID:            uuid.New(),
		BookingID:     bookingID,
		TransactionID: "tx-1",
		Status:        model.PaymentStatusPending,
	}

	paymentRepo := new(paymentRepositoryMock)
	bookingRepo := new(bookingRepositoryMock)
	bookingSvc := new(bookingServiceInterfaceMock)
	svc := NewPaymentService(paymentRepo, bookingRepo, bookingSvc)

	paymentRepo.On("GetByTransactionID", ctx, "tx-1").Return(payment, nil).Once()
	paymentRepo.On("Update", ctx, mock.MatchedBy(func(updated *model.Payment) bool {
		return updated.Status == model.PaymentStatusSuccess && updated.PaidAt != nil
	})).Return(nil).Once()
	bookingSvc.On("ConfirmBooking", ctx, bookingID).Return(nil).Once()

	err := svc.ProcessWebhook(ctx, &model.WebhookRequest{
		TransactionID: "tx-1",
		Status:        "success",
		Amount:        100,
	})

	require.NoError(t, err)
	paymentRepo.AssertExpectations(t)
	bookingSvc.AssertExpectations(t)
	bookingRepo.AssertExpectations(t)
}

func TestPaymentServiceGetPaymentByBooking(t *testing.T) {
	ctx := context.Background()
	bookingID := uuid.New()
	now := time.Now()

	paymentRepo := new(paymentRepositoryMock)
	bookingRepo := new(bookingRepositoryMock)
	bookingSvc := new(bookingServiceInterfaceMock)
	svc := NewPaymentService(paymentRepo, bookingRepo, bookingSvc)

	paymentRepo.On("GetByBookingID", ctx, bookingID).Return(&model.Payment{
		ID:            uuid.New(),
		BookingID:     bookingID,
		Amount:        320,
		Status:        model.PaymentStatusSuccess,
		PaymentMethod: model.PaymentMethodSBP,
		PaidAt:        &now,
	}, nil).Once()

	resp, err := svc.GetPaymentByBooking(ctx, bookingID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, bookingID, resp.BookingID)
	assert.Equal(t, 320.0, resp.Amount)
	assert.Equal(t, model.PaymentStatusSuccess, resp.Status)
	assert.Equal(t, model.PaymentMethodSBP, resp.PaymentMethod)

	paymentRepo.AssertExpectations(t)
}

func TestPaymentServiceGetPaymentSuccess(t *testing.T) {
	ctx := context.Background()
	paymentID := uuid.New()
	bookingID := uuid.New()

	paymentRepo := new(paymentRepositoryMock)
	svc := NewPaymentService(paymentRepo, new(bookingRepositoryMock), new(bookingServiceInterfaceMock))

	paymentRepo.On("GetByID", ctx, paymentID).Return(&model.Payment{
		ID:        paymentID,
		BookingID: bookingID,
		Amount:    150,
		Status:    model.PaymentStatusPending,
	}, nil).Once()

	resp, err := svc.GetPayment(ctx, paymentID)
	assert.NoError(t, err)
	assert.Equal(t, paymentID, resp.ID)
	assert.Equal(t, bookingID, resp.BookingID)
}
