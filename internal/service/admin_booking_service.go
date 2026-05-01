package service

import (
	"context"
	"errors"
	"time"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/google/uuid"
)

type AdminBookingService struct {
	bookingRepo repository.AdminBookingRepositoryInterface
	seatRepo    repository.SeatRepositoryInterface
	eventRepo   repository.EventRepositoryInterface
}

func NewAdminBookingService(
	bookingRepo repository.AdminBookingRepositoryInterface,
	seatRepo repository.SeatRepositoryInterface,
	eventRepo repository.EventRepositoryInterface,
) *AdminBookingService {
	return &AdminBookingService{
		bookingRepo: bookingRepo,
		seatRepo:    seatRepo,
		eventRepo:   eventRepo,
	}
}

func (s *AdminBookingService) GetAllBookings(ctx context.Context, page, limit int, status, userID, eventID string) ([]*model.Booking, int, error) {
	return s.bookingRepo.GetAllBookings(ctx, page, limit, status, userID, eventID)
}

func (s *AdminBookingService) GetBookingDetail(ctx context.Context, id uuid.UUID) (*model.BookingDetail, error) {
	booking, err := s.bookingRepo.GetBookingByID(ctx, id)
	if err != nil {
		return nil, err
	}

	seats, _ := s.bookingRepo.GetBookingSeats(ctx, id)
	payment, _ := s.bookingRepo.GetBookingPayment(ctx, id)

	detail := &model.BookingDetail{
		Booking: *booking,
		Seats:   seats,
		Payment: payment,
	}

	return detail, nil
}

func (s *AdminBookingService) CancelBooking(ctx context.Context, id uuid.UUID, refund bool) error {
	booking, err := s.bookingRepo.GetBookingByID(ctx, id)
	if err != nil {
		return err
	}

	if booking.Status == model.BookingStatusCancelled || booking.Status == model.BookingStatusCompleted {
		return errors.New("cannot cancel this booking")
	}

	seats, err := s.bookingRepo.GetBookingSeats(ctx, id)
	if err != nil {
		return err
	}

	for _, bs := range seats {
		seat, err := s.seatRepo.GetByID(ctx, bs.SeatID)
		if err != nil {
			continue
		}
		if seat.Status == model.SeatStatusBooked || seat.Status == model.SeatStatusReserved {
			seat.Status = model.SeatStatusAvailable
			if err := s.seatRepo.Update(ctx, seat); err != nil {
				return err
			}
		}
	}

	event, err := s.eventRepo.GetByID(ctx, booking.EventID)
	if err == nil {
		event.AvailableSeats += len(seats)
		if err := s.eventRepo.Update(ctx, event); err != nil {
			return err
		}
	}

	if err := s.bookingRepo.CancelBookingAdmin(ctx, id, model.BookingStatusCancelled); err != nil {
		return err
	}

	if refund {
		payment, err := s.bookingRepo.GetBookingPayment(ctx, id)
		if err == nil && payment != nil {
			if err := s.bookingRepo.RefundPayment(ctx, payment.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *AdminBookingService) RefundBooking(ctx context.Context, id uuid.UUID) error {
	payment, err := s.bookingRepo.GetBookingPayment(ctx, id)
	if err != nil {
		return errors.New("payment not found")
	}

	if payment.Status == model.PaymentStatusRefunded {
		return errors.New("payment already refunded")
	}

	return s.bookingRepo.RefundPayment(ctx, payment.ID)
}

func (s *AdminBookingService) GetBookingsStats(ctx context.Context) (*model.BookingsStats, error) {
	bookingStats, _ := s.bookingRepo.GetBookingsStats(ctx)
	revenueStats, _ := s.bookingRepo.GetRevenueStats(ctx)

	return &model.BookingsStats{
		TotalBookings:     bookingStats["total"],
		PendingBookings:   bookingStats["pending"],
		ConfirmedBookings: bookingStats["confirmed"],
		CancelledBookings: bookingStats["cancelled"],
		TotalRevenue:      revenueStats["total"],
	}, nil
}

func (s *AdminBookingService) ExportBookingsToCSV(ctx context.Context, status string) ([][]string, error) {
	bookings, _, err := s.bookingRepo.GetAllBookings(ctx, 1, 10000, status, "", "")
	if err != nil {
		return nil, err
	}

	var rows [][]string

	rows = append(rows, []string{
		"Booking ID", "User Email", "Event Title", "Event Date",
		"Seats", "Total Amount", "Discount", "Final Amount",
		"Status", "Created At",
	})

	for _, b := range bookings {
		seats, _ := s.bookingRepo.GetBookingSeats(ctx, b.ID)
		seatNumbers := ""
		for i, s := range seats {
			if i > 0 {
				seatNumbers += ", "
			}
			seatNumbers += s.RowNumber + "-" + s.SeatNumber
		}

		rows = append(rows, []string{
			b.ID.String(),
			b.Email,
			b.EventTitle,
			b.EventDate.Format(time.RFC3339),
			seatNumbers,
			formatMoney(b.TotalAmount),
			formatMoney(b.Discount),
			formatMoney(b.FinalAmount),
			string(b.Status),
			b.CreatedAt.Format(time.RFC3339),
		})
	}

	return rows, nil
}

func formatMoney(amount float64) string {
	return string(rune(amount))
}
