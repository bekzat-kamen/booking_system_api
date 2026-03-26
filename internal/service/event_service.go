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
	ErrEventForbidden     = errors.New("event access forbidden")
	ErrInvalidEventDate   = errors.New("invalid event date format")
	ErrInvalidSeatCount   = errors.New("invalid total seats value")
	ErrInvalidSeatBalance = errors.New("total seats cannot be less than already booked seats")
)

type EventService struct {
	eventRepo *repository.EventRepository
}

func NewEventService(eventRepo *repository.EventRepository) *EventService {
	return &EventService{eventRepo: eventRepo}
}

func (s *EventService) Create(ctx context.Context, organizerID uuid.UUID, req *model.CreateEventRequest) (*model.Event, error) {
	eventDate, err := time.Parse(time.RFC3339, req.EventDate)
	if err != nil {
		return nil, ErrInvalidEventDate
	}

	event := &model.Event{
		ID:             uuid.New(),
		Title:          req.Title,
		Description:    req.Description,
		Venue:          req.Venue,
		EventDate:      eventDate,
		TotalSeats:     req.TotalSeats,
		AvailableSeats: req.TotalSeats,
		Price:          req.Price,
		Status:         model.EventStatusDraft,
		CreatedBy:      organizerID,
		ImageURL:       req.ImageURL,
	}

	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, err
	}

	return s.eventRepo.GetByID(ctx, event.ID)
}

func (s *EventService) GetByID(ctx context.Context, eventID uuid.UUID) (*model.Event, error) {
	return s.eventRepo.GetByID(ctx, eventID)
}

func (s *EventService) GetAll(ctx context.Context, limit, offset int) ([]*model.Event, int, error) {
	events, err := s.eventRepo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.eventRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (s *EventService) GetByOrganizer(ctx context.Context, organizerID uuid.UUID, limit, offset int) ([]*model.Event, error) {
	return s.eventRepo.GetByOrganizer(ctx, organizerID, limit, offset)
}

func (s *EventService) Update(ctx context.Context, eventID, organizerID uuid.UUID, req *model.UpdateEventRequest) (*model.Event, error) {
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	if event.CreatedBy != organizerID {
		return nil, ErrEventForbidden
	}

	if req.Title != "" {
		event.Title = req.Title
	}
	if req.Description != "" {
		event.Description = req.Description
	}
	if req.Venue != "" {
		event.Venue = req.Venue
	}
	if req.EventDate != "" {
		eventDate, err := time.Parse(time.RFC3339, req.EventDate)
		if err != nil {
			return nil, ErrInvalidEventDate
		}
		event.EventDate = eventDate
	}
	if req.TotalSeats < 0 {
		return nil, ErrInvalidSeatCount
	}
	if req.TotalSeats > 0 {
		bookedSeats := event.TotalSeats - event.AvailableSeats
		if req.TotalSeats < bookedSeats {
			return nil, ErrInvalidSeatBalance
		}
		event.TotalSeats = req.TotalSeats
		event.AvailableSeats = req.TotalSeats - bookedSeats
	}
	if req.Price > 0 {
		event.Price = req.Price
	}
	if req.Status != "" {
		event.Status = model.EventStatus(req.Status)
	}
	if req.ImageURL != "" {
		event.ImageURL = req.ImageURL
	}

	if err := s.eventRepo.Update(ctx, event); err != nil {
		return nil, err
	}

	return s.eventRepo.GetByID(ctx, event.ID)
}

func (s *EventService) Delete(ctx context.Context, eventID, organizerID uuid.UUID) error {
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return err
	}

	if event.CreatedBy != organizerID {
		return ErrEventForbidden
	}

	return s.eventRepo.Delete(ctx, eventID)
}

func (s *EventService) PublishEvent(ctx context.Context, eventID, organizerID uuid.UUID) (*model.Event, error) {
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	if event.CreatedBy != organizerID {
		return nil, ErrEventForbidden
	}

	event.Status = model.EventStatusPublished
	if event.AvailableSeats <= 0 {
		event.Status = model.EventStatusSoldOut
	}

	if err := s.eventRepo.Update(ctx, event); err != nil {
		return nil, err
	}

	return s.eventRepo.GetByID(ctx, event.ID)
}
