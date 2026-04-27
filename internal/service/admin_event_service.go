package service

import (
	"context"
	"errors"
	"time"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/google/uuid"
)

type AdminEventService struct {
	eventRepo repository.AdminEventRepositoryInterface
}

func NewAdminEventService(eventRepo repository.AdminEventRepositoryInterface) *AdminEventService {
	return &AdminEventService{eventRepo: eventRepo}
}

func (s *AdminEventService) GetAllEvents(ctx context.Context, page, limit int, status, organizerID string) ([]*model.Event, int, error) {
	return s.eventRepo.GetAllEvents(ctx, page, limit, status, organizerID)
}

func (s *AdminEventService) GetEventDetail(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) {
	event, err := s.eventRepo.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}

	stats, _ := s.eventRepo.GetEventStats(ctx, id)

	detail := map[string]interface{}{
		"event": map[string]interface{}{
			"id":              event.ID,
			"title":           event.Title,
			"description":     event.Description,
			"venue":           event.Venue,
			"event_date":      event.EventDate,
			"total_seats":     event.TotalSeats,
			"available_seats": event.AvailableSeats,
			"price":           event.Price,
			"status":          event.Status,
			"created_by":      event.CreatedBy,
			"organizer_email": event.ImageURL,
			"image_url":       event.ImageURL,
			"created_at":      event.CreatedAt,
			"updated_at":      event.UpdatedAt,
		},
		"statistics": stats,
	}

	return detail, nil
}

func (s *AdminEventService) UpdateEvent(ctx context.Context, id uuid.UUID, req *model.UpdateEventRequest) (*model.Event, error) {
	event, err := s.eventRepo.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
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
			return nil, errors.New("invalid event date format")
		}
		event.EventDate = eventDate
	}
	if req.TotalSeats > 0 {
		event.TotalSeats = req.TotalSeats
		if event.AvailableSeats > event.TotalSeats {
			event.AvailableSeats = event.TotalSeats
		}
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

	if err := s.eventRepo.UpdateEventAdmin(ctx, event); err != nil {
		return nil, errors.New("failed to update event")
	}

	return event, nil
}

func (s *AdminEventService) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	return s.eventRepo.DeleteEventAdmin(ctx, id)
}

func (s *AdminEventService) PublishEvent(ctx context.Context, id uuid.UUID) (*model.Event, error) {
	if err := s.eventRepo.PublishEventAdmin(ctx, id); err != nil {
		return nil, err
	}
	return s.eventRepo.GetEventByID(ctx, id)
}

func (s *AdminEventService) GetEventsStats(ctx context.Context) (map[string]int64, error) {
	return s.eventRepo.GetEventsByStatus(ctx)
}
