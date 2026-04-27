package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrPaymentAlreadyExists = errors.New("payment already exists for this booking")
	ErrInvalidPaymentStatus = errors.New("invalid payment status")
)

type PaymentService struct {
	paymentRepo    repository.PaymentRepositoryInterface
	bookingRepo    repository.BookingRepositoryInterface
	bookingService BookingServiceInterface
}

func NewPaymentService(
	paymentRepo repository.PaymentRepositoryInterface,
	bookingRepo repository.BookingRepositoryInterface,
	bookingService BookingServiceInterface,
) *PaymentService {
	return &PaymentService{
		paymentRepo:    paymentRepo,
		bookingRepo:    bookingRepo,
		bookingService: bookingService,
	}
}

func (s *PaymentService) CreatePayment(ctx context.Context, req *model.CreatePaymentRequest) (*model.Payment, error) {

	booking, err := s.bookingRepo.GetByID(ctx, req.BookingID)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			return nil, repository.ErrBookingNotFound
		}
		return nil, err
	}

	if booking.Status != model.BookingStatusPending {
		return nil, errors.New("booking is not in pending status")
	}

	_, err = s.paymentRepo.GetByBookingID(ctx, req.BookingID)
	if err == nil {
		return nil, ErrPaymentAlreadyExists
	}

	payment := &model.Payment{
		ID:            uuid.New(),
		BookingID:     req.BookingID,
		TransactionID: generateMockTransactionID(),
		Amount:        booking.FinalAmount,
		Status:        model.PaymentStatusPending,
		PaymentMethod: req.PaymentMethod,
		Provider:      "mock", // Для имитации
		PaidAt:        nil,
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, errors.New("failed to create payment")
	}

	return payment, nil
}

func (s *PaymentService) ProcessPayment(ctx context.Context, paymentID uuid.UUID) (*model.Payment, error) {

	payment, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	if payment.Status != model.PaymentStatusPending {
		return nil, ErrInvalidPaymentStatus
	}

	now := time.Now()
	payment.Status = model.PaymentStatusSuccess
	payment.PaidAt = &now
	payment.ProviderResponse = `{"mock": true, "status": "success", "message": "Payment simulated"}`

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, errors.New("failed to update payment")
	}

	if err := s.bookingService.ConfirmBooking(ctx, payment.BookingID); err != nil {
		payment.Status = model.PaymentStatusFailed
		s.paymentRepo.Update(ctx, payment)
		return nil, errors.New("failed to confirm booking")
	}

	return payment, nil
}

func (s *PaymentService) ProcessWebhook(ctx context.Context, req *model.WebhookRequest) error {
	payment, err := s.paymentRepo.GetByTransactionID(ctx, req.TransactionID)
	if err != nil {
		return err
	}

	// 2. Проверяем подпись (в реальности)
	// if !verifySignature(req.Signature, payment) {
	//     return errors.New("invalid signature")
	// }

	var newStatus model.PaymentStatus
	switch req.Status {
	case "success", "succeeded":
		newStatus = model.PaymentStatusSuccess
	case "failed", "error":
		newStatus = model.PaymentStatusFailed
	case "refunded":
		newStatus = model.PaymentStatusRefunded
	default:
		return ErrInvalidPaymentStatus
	}

	payment.Status = newStatus
	if newStatus == model.PaymentStatusSuccess {
		now := time.Now()
		payment.PaidAt = &now
	}

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return err
	}

	if newStatus == model.PaymentStatusSuccess {
		return s.bookingService.ConfirmBooking(ctx, payment.BookingID)
	}

	return nil
}

func (s *PaymentService) GetPayment(ctx context.Context, id uuid.UUID) (*model.PaymentResponse, error) {
	payment, err := s.paymentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &model.PaymentResponse{
		ID:            payment.ID,
		BookingID:     payment.BookingID,
		Amount:        payment.Amount,
		Status:        payment.Status,
		PaymentMethod: payment.PaymentMethod,
		PaidAt:        payment.PaidAt,
	}, nil
}

func (s *PaymentService) GetPaymentByBooking(ctx context.Context, bookingID uuid.UUID) (*model.PaymentResponse, error) {
	payment, err := s.paymentRepo.GetByBookingID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	return &model.PaymentResponse{
		ID:            payment.ID,
		BookingID:     payment.BookingID,
		Amount:        payment.Amount,
		Status:        payment.Status,
		PaymentMethod: payment.PaymentMethod,
		PaidAt:        payment.PaidAt,
	}, nil
}
func generateMockTransactionID() string {
	return fmt.Sprintf("MOCK_%d", time.Now().UnixNano())
}
