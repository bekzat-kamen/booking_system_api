package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrSeatsAlreadyGenerated = errors.New("seats already generated for this event")
	ErrInvalidSeatConfig     = errors.New("invalid seat configuration")
)

type SeatService struct {
	seatRepo  *repository.SeatRepository
	eventRepo *repository.EventRepository
}

func NewSeatService(seatRepo *repository.SeatRepository, eventRepo *repository.EventRepository) *SeatService {
	return &SeatService{
		seatRepo:  seatRepo,
		eventRepo: eventRepo,
	}
}

func (s *SeatService) GenerateSeats(ctx context.Context, eventID, organizerID uuid.UUID, req *model.GenerateSeatsRequest) error {
	// 1. Проверяем существование мероприятия
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			return repository.ErrEventNotFound
		}
		return err
	}
	if event.CreatedBy != organizerID {
		return ErrEventForbidden
	}

	// 2. Проверяем, не сгенерированы ли уже места
	existingSeats, err := s.seatRepo.GetByEvent(ctx, eventID)
	if err != nil {
		return err
	}
	if len(existingSeats) > 0 {
		return ErrSeatsAlreadyGenerated
	}

	// 3. Валидация конфигурации
	if req.TotalRows < 1 || req.TotalRows > 50 {
		return ErrInvalidSeatConfig
	}
	if req.SeatsPerRow < 1 || req.SeatsPerRow > 100 {
		return ErrInvalidSeatConfig
	}

	// 4. Генерируем места
	var seats []*model.Seat
	totalSeats := req.TotalRows * req.SeatsPerRow

	for row := 1; row <= req.TotalRows; row++ {
		for seat := 1; seat <= req.SeatsPerRow; seat++ {
			seats = append(seats, &model.Seat{
				ID:         uuid.New(),
				EventID:    eventID,
				SeatNumber: fmt.Sprintf("%d", seat),
				RowNumber:  fmt.Sprintf("%d", row),
				Status:     model.SeatStatusAvailable,
				Price:      req.BasePrice,
				Version:    0,
			})
		}
	}

	// 5. Сохраняем в БД (транзакция)
	if err := s.seatRepo.CreateBatch(ctx, seats); err != nil {
		return errors.New("failed to generate seats")
	}

	// 6. Обновляем количество мест в мероприятии
	event.TotalSeats = totalSeats
	event.AvailableSeats = totalSeats
	if err := s.eventRepo.Update(ctx, event); err != nil {
		return errors.New("failed to update event seats count")
	}

	return nil
}

func (s *SeatService) GetSeatMap(ctx context.Context, eventID uuid.UUID) (*model.SeatMapResponse, error) {
	// 1. Проверяем существование мероприятия
	_, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			return nil, repository.ErrEventNotFound
		}
		return nil, err
	}

	// 2. Получаем все места
	seats, err := s.seatRepo.GetByEvent(ctx, eventID)
	if err != nil {
		return nil, errors.New("failed to get seats")
	}

	// 3. Считаем статистику
	available := 0
	booked := 0
	reserved := 0
	blocked := 0

	seatResponses := make([]model.SeatResponse, 0, len(seats))
	for _, seat := range seats {
		switch seat.Status {
		case model.SeatStatusAvailable:
			available++
		case model.SeatStatusBooked:
			booked++
		case model.SeatStatusReserved:
			reserved++
		case model.SeatStatusBlocked:
			blocked++
		}

		seatResponses = append(seatResponses, model.SeatResponse{
			ID:         seat.ID,
			EventID:    seat.EventID,
			SeatNumber: seat.SeatNumber,
			RowNumber:  seat.RowNumber,
			Status:     seat.Status,
			Price:      seat.Price,
		})
	}

	return &model.SeatMapResponse{
		EventID:    eventID,
		TotalSeats: len(seats),
		Available:  available,
		Booked:     booked,
		Reserved:   reserved,
		Blocked:    blocked,
		Seats:      seatResponses,
	}, nil
}

func (s *SeatService) GetAvailableSeats(ctx context.Context, eventID uuid.UUID) ([]model.SeatResponse, error) {
	seats, err := s.seatRepo.GetByEventAndStatus(ctx, eventID, model.SeatStatusAvailable)
	if err != nil {
		return nil, err
	}

	responses := make([]model.SeatResponse, 0, len(seats))
	for _, seat := range seats {
		responses = append(responses, model.SeatResponse{
			ID:         seat.ID,
			EventID:    seat.EventID,
			SeatNumber: seat.SeatNumber,
			RowNumber:  seat.RowNumber,
			Status:     seat.Status,
			Price:      seat.Price,
		})
	}

	return responses, nil
}

func (s *SeatService) ReserveSeat(ctx context.Context, seatID uuid.UUID) error {
	seat, err := s.seatRepo.GetByID(ctx, seatID)
	if err != nil {
		return err
	}

	if seat.Status != model.SeatStatusAvailable {
		return repository.ErrSeatNotAvailable
	}

	seat.Status = model.SeatStatusReserved
	return s.seatRepo.Update(ctx, seat)
}

func (s *SeatService) BookSeat(ctx context.Context, seatID uuid.UUID) error {
	seat, err := s.seatRepo.GetByID(ctx, seatID)
	if err != nil {
		return err
	}

	if seat.Status != model.SeatStatusReserved && seat.Status != model.SeatStatusAvailable {
		return repository.ErrSeatNotAvailable
	}

	seat.Status = model.SeatStatusBooked
	return s.seatRepo.Update(ctx, seat)
}

func (s *SeatService) ReleaseSeat(ctx context.Context, seatID uuid.UUID) error {
	seat, err := s.seatRepo.GetByID(ctx, seatID)
	if err != nil {
		return err
	}

	if seat.Status == model.SeatStatusReserved {
		seat.Status = model.SeatStatusAvailable
		return s.seatRepo.Update(ctx, seat)
	}

	return nil
}
