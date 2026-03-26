package service

import (
	"context"
	"errors"
	"time"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrSeatsNotAvailable = errors.New("one or more seats are not available")
	ErrBookingExpired    = errors.New("booking has expired")
	ErrCannotCancel      = errors.New("cannot cancel this booking")
)

type BookingService struct {
	bookingRepo *repository.BookingRepository
	seatRepo    *repository.SeatRepository
	eventRepo   *repository.EventRepository
}

func NewBookingService(
	bookingRepo *repository.BookingRepository,
	seatRepo *repository.SeatRepository,
	eventRepo *repository.EventRepository,
) *BookingService {
	return &BookingService{
		bookingRepo: bookingRepo,
		seatRepo:    seatRepo,
		eventRepo:   eventRepo,
	}
}

func (s *BookingService) CreateBooking(ctx context.Context, userID uuid.UUID, req *model.CreateBookingRequest) (*model.BookingResponse, error) {
	event, err := s.eventRepo.GetByID(ctx, req.EventID)
	if err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			return nil, repository.ErrEventNotFound
		}
		return nil, err
	}

	seats := make([]*model.Seat, 0, len(req.SeatIDs))
	totalAmount := 0.0

	for _, seatID := range req.SeatIDs {
		seat, err := s.seatRepo.GetByID(ctx, seatID)
		if err != nil {
			return nil, repository.ErrSeatNotFound
		}

		if seat.EventID != req.EventID {
			return nil, errors.New("seat does not belong to this event")
		}

		if seat.Status != model.SeatStatusAvailable {
			return nil, ErrSeatsNotAvailable
		}

		seats = append(seats, seat)
		totalAmount += seat.Price
	}

	if event.AvailableSeats < len(req.SeatIDs) {
		return nil, errors.New("not enough available seats")
	}

	booking := &model.Booking{
		ID:          uuid.New(),
		UserID:      userID,
		EventID:     req.EventID,
		TotalAmount: totalAmount,
		Discount:    0,
		FinalAmount: totalAmount,
		Status:      model.BookingStatusPending,
		ExpiresAt:   time.Now().Add(15 * time.Minute),
	}

	bookingSeats := make([]*model.BookingSeat, 0, len(req.SeatIDs))
	for _, seat := range seats {
		bookingSeats = append(bookingSeats, &model.BookingSeat{
			BookingID: booking.ID,
			SeatID:    seat.ID,
			Price:     seat.Price,
		})
	}

	if err := s.bookingRepo.Create(ctx, booking, bookingSeats); err != nil {
		return nil, errors.New("failed to create booking")
	}

	for _, seat := range seats {
		seat.Status = model.SeatStatusReserved
		if err := s.seatRepo.Update(ctx, seat); err != nil {

			s.bookingRepo.UpdateStatus(ctx, booking.ID, model.BookingStatusCancelled)
			return nil, errors.New("failed to reserve seats")
		}
	}

	event.AvailableSeats -= len(req.SeatIDs)
	if err := s.eventRepo.Update(ctx, event); err != nil {
		return nil, errors.New("failed to update event seats")
	}

	return s.getBookingWithSeats(ctx, booking.ID)
}

func (s *BookingService) GetBooking(ctx context.Context, id uuid.UUID) (*model.BookingResponse, error) {
	booking, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.getBookingWithSeats(ctx, booking.ID)
}

func (s *BookingService) GetUserBookings(ctx context.Context, userID uuid.UUID, page, limit int) ([]*model.BookingResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	bookings, err := s.bookingRepo.GetByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.bookingRepo.CountByUser(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*model.BookingResponse, 0, len(bookings))
	for _, booking := range bookings {
		response, err := s.getBookingWithSeats(ctx, booking.ID)
		if err != nil {
			continue
		}
		responses = append(responses, response)
	}

	return responses, total, nil
}

func (s *BookingService) CancelBooking(ctx context.Context, bookingID uuid.UUID, userID uuid.UUID) error {

	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.UserID != userID {
		return errors.New("unauthorized to cancel this booking")
	}

	if booking.Status != model.BookingStatusPending && booking.Status != model.BookingStatusConfirmed {
		return ErrCannotCancel
	}

	seats, err := s.bookingRepo.GetSeats(ctx, bookingID)
	if err != nil {
		return err
	}

	for _, bs := range seats {
		seat, err := s.seatRepo.GetByID(ctx, bs.SeatID)
		if err != nil {
			continue
		}
		seat.Status = model.SeatStatusAvailable
		s.seatRepo.Update(ctx, seat)
	}

	booking.Status = model.BookingStatusCancelled
	if err := s.bookingRepo.Update(ctx, booking); err != nil {
		return err
	}

	event, err := s.eventRepo.GetByID(ctx, booking.EventID)
	if err == nil {
		event.AvailableSeats += len(seats)
		s.eventRepo.Update(ctx, event)
	}

	return nil
}

func (s *BookingService) ConfirmBooking(ctx context.Context, bookingID uuid.UUID) error {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.Status != model.BookingStatusPending {
		return errors.New("booking is not in pending status")
	}

	if time.Now().After(booking.ExpiresAt) {
		return ErrBookingExpired
	}

	now := time.Now()
	booking.Status = model.BookingStatusConfirmed
	booking.PaidAt = &now

	return s.bookingRepo.Update(ctx, booking)
}

func (s *BookingService) getBookingWithSeats(ctx context.Context, bookingID uuid.UUID) (*model.BookingResponse, error) {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	bookingSeats, err := s.bookingRepo.GetSeats(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	seats := make([]model.SeatResponse, 0, len(bookingSeats))
	for _, bs := range bookingSeats {
		seat, err := s.seatRepo.GetByID(ctx, bs.SeatID)
		if err != nil {
			continue
		}
		seats = append(seats, model.SeatResponse{
			ID:         seat.ID,
			EventID:    seat.EventID,
			SeatNumber: seat.SeatNumber,
			RowNumber:  seat.RowNumber,
			Status:     seat.Status,
			Price:      seat.Price,
		})
	}

	event, err := s.eventRepo.GetByID(ctx, booking.EventID)
	eventTitle := ""
	if err == nil {
		eventTitle = event.Title
	}

	return &model.BookingResponse{
		ID:          booking.ID,
		UserID:      booking.UserID,
		EventID:     booking.EventID,
		EventTitle:  eventTitle,
		TotalAmount: booking.TotalAmount,
		Discount:    booking.Discount,
		FinalAmount: booking.FinalAmount,
		Status:      booking.Status,
		ExpiresAt:   booking.ExpiresAt,
		PaidAt:      booking.PaidAt,
		Seats:       seats,
		CreatedAt:   booking.CreatedAt,
	}, nil
}

func (s *BookingService) ExpirePendingBookings(ctx context.Context) (int, error) {
	expiredBookings, err := s.bookingRepo.GetExpiredPending(ctx)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, booking := range expiredBookings {

		seats, err := s.bookingRepo.GetSeats(ctx, booking.ID)
		if err != nil {
			continue
		}

		for _, bs := range seats {
			seat, err := s.seatRepo.GetByID(ctx, bs.SeatID)
			if err != nil {
				continue
			}
			if seat.Status == model.SeatStatusReserved {
				seat.Status = model.SeatStatusAvailable
				s.seatRepo.Update(ctx, seat)
			}
		}

		booking.Status = model.BookingStatusExpired
		s.bookingRepo.Update(ctx, booking)

		event, err := s.eventRepo.GetByID(ctx, booking.EventID)
		if err == nil {
			event.AvailableSeats += len(seats)
			s.eventRepo.Update(ctx, event)
		}

		count++
	}

	return count, nil
}
