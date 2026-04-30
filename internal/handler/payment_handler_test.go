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

type paymentServiceMock struct {
	mock.Mock
}

func (m *paymentServiceMock) CreatePayment(ctx context.Context, req *model.CreatePaymentRequest) (*model.Payment, error) {
	args := m.Called(ctx, req)
	payment, _ := args.Get(0).(*model.Payment)
	return payment, args.Error(1)
}

func (m *paymentServiceMock) ProcessPayment(ctx context.Context, paymentID uuid.UUID) (*model.Payment, error) {
	args := m.Called(ctx, paymentID)
	payment, _ := args.Get(0).(*model.Payment)
	return payment, args.Error(1)
}

func (m *paymentServiceMock) ProcessWebhook(ctx context.Context, req *model.WebhookRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *paymentServiceMock) GetPayment(ctx context.Context, id uuid.UUID) (*model.PaymentResponse, error) {
	args := m.Called(ctx, id)
	resp, _ := args.Get(0).(*model.PaymentResponse)
	return resp, args.Error(1)
}

func (m *paymentServiceMock) GetPaymentByBooking(ctx context.Context, bookingID uuid.UUID) (*model.PaymentResponse, error) {
	args := m.Called(ctx, bookingID)
	resp, _ := args.Get(0).(*model.PaymentResponse)
	return resp, args.Error(1)
}

func TestPaymentHandlerCreatePaymentSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	paymentSvc := new(paymentServiceMock)
	h := NewPaymentHandler(paymentSvc)
	router := gin.New()
	router.POST("/payments", h.CreatePayment)

	bookingID := uuid.New()
	paymentSvc.On("CreatePayment", mock.Anything, mock.MatchedBy(func(req *model.CreatePaymentRequest) bool {
		return req.BookingID == bookingID && req.PaymentMethod == model.PaymentMethodCard
	})).Return(&model.Payment{
		ID:            uuid.New(),
		BookingID:     bookingID,
		Amount:        100,
		Status:        model.PaymentStatusPending,
		PaymentMethod: model.PaymentMethodCard,
	}, nil).Once()

	w := performJSONRequest(router, http.MethodPost, "/payments", map[string]string{
		"booking_id":     bookingID.String(),
		"payment_method": "card",
	})

	require.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), bookingID.String())
	paymentSvc.AssertExpectations(t)
}

func TestPaymentHandlerCreatePaymentConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)

	paymentSvc := new(paymentServiceMock)
	h := NewPaymentHandler(paymentSvc)
	router := gin.New()
	router.POST("/payments", h.CreatePayment)

	bookingID := uuid.New()
	paymentSvc.On("CreatePayment", mock.Anything, mock.AnythingOfType("*model.CreatePaymentRequest")).
		Return((*model.Payment)(nil), service.ErrPaymentAlreadyExists).Once()

	w := performJSONRequest(router, http.MethodPost, "/payments", map[string]string{
		"booking_id":     bookingID.String(),
		"payment_method": "card",
	})

	require.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), service.ErrPaymentAlreadyExists.Error())
	paymentSvc.AssertExpectations(t)
}

func TestPaymentHandlerProcessPaymentBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	paymentSvc := new(paymentServiceMock)
	h := NewPaymentHandler(paymentSvc)
	router := gin.New()
	router.POST("/payments/:id/process", h.ProcessPayment)

	w := performJSONRequest(router, http.MethodPost, "/payments/not-a-uuid/process", nil)

	require.Equal(t, http.StatusBadRequest, w.Code)
	paymentSvc.AssertExpectations(t)
}

func TestPaymentHandlerGetPaymentSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	paymentSvc := new(paymentServiceMock)
	h := NewPaymentHandler(paymentSvc)
	router := gin.New()
	router.GET("/payments/:id", h.GetPayment)

	paymentID := uuid.New()
	paymentSvc.On("GetPayment", mock.Anything, paymentID).Return(&model.PaymentResponse{
		ID:            paymentID,
		BookingID:     uuid.New(),
		Amount:        100,
		Status:        model.PaymentStatusSuccess,
		PaymentMethod: model.PaymentMethodCard,
	}, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/payments/"+paymentID.String(), nil)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), paymentID.String())
	paymentSvc.AssertExpectations(t)
}

func TestPaymentHandlerGetPaymentByBookingNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	paymentSvc := new(paymentServiceMock)
	h := NewPaymentHandler(paymentSvc)
	router := gin.New()
	router.GET("/payments/booking/:booking_id", h.GetPaymentByBooking)

	bookingID := uuid.New()
	paymentSvc.On("GetPaymentByBooking", mock.Anything, bookingID).
		Return((*model.PaymentResponse)(nil), repository.ErrPaymentNotFound).Once()

	w := performJSONRequest(router, http.MethodGet, "/payments/booking/"+bookingID.String(), nil)

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "payment not found")
	paymentSvc.AssertExpectations(t)
}

func TestPaymentHandlerWebhookSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	paymentSvc := new(paymentServiceMock)
	h := NewPaymentHandler(paymentSvc)
	router := gin.New()
	router.POST("/payments/webhook", h.Webhook)

	paymentSvc.On("ProcessWebhook", mock.Anything, mock.MatchedBy(func(req *model.WebhookRequest) bool {
		return req.TransactionID == "tx-1" && req.Status == "success"
	})).Return(nil).Once()

	w := performJSONRequest(router, http.MethodPost, "/payments/webhook", map[string]interface{}{
		"transaction_id": "tx-1",
		"status":         "success",
		"amount":         100,
		"signature":      "sig",
	})

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "webhook processed successfully")
	paymentSvc.AssertExpectations(t)
}
