package model

import (
	"time"

	"github.com/google/uuid"
)

type EventStatus string

const (
	EventStatusDraft     EventStatus = "draft"
	EventStatusPublished EventStatus = "published"
	EventStatusSoldOut   EventStatus = "sold_out"
	EventStatusCancelled EventStatus = "cancelled"
	EventStatusCompleted EventStatus = "completed"
)

type Event struct {
	ID             uuid.UUID   `db:"id" json:"id"`
	Title          string      `db:"title" json:"title"`
	Description    string      `db:"description" json:"description"`
	Venue          string      `db:"venue" json:"venue"`
	EventDate      time.Time   `db:"event_date" json:"event_date"`
	TotalSeats     int         `db:"total_seats" json:"total_seats"`
	AvailableSeats int         `db:"available_seats" json:"available_seats"`
	Price          float64     `db:"price" json:"price"`
	Status         EventStatus `db:"status" json:"status"`
	CreatedBy      uuid.UUID   `db:"created_by" json:"created_by"`
	OrganizerEmail string      `db:"organizer_email" json:"organizer_email,omitempty"`
	ImageURL       string      `db:"image_url" json:"image_url"`
	CreatedAt      time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time   `db:"updated_at" json:"updated_at"`
}

type CreateEventRequest struct {
	Title       string  `json:"title" binding:"required,min=3,max=255"`
	Description string  `json:"description" binding:"omitempty,max=5000"`
	Venue       string  `json:"venue" binding:"required,min=3,max=255"`
	EventDate   string  `json:"event_date" binding:"required"` // RFC3339 формат
	TotalSeats  int     `json:"total_seats" binding:"required,min=1"`
	Price       float64 `json:"price" binding:"required,min=0"`
	ImageURL    string  `json:"image_url" binding:"omitempty,url"`
}

type UpdateEventRequest struct {
	Title       string  `json:"title" binding:"omitempty,min=3,max=255"`
	Description string  `json:"description" binding:"omitempty,max=5000"`
	Venue       string  `json:"venue" binding:"omitempty,min=3,max=255"`
	EventDate   string  `json:"event_date" binding:"omitempty"`
	TotalSeats  int     `json:"total_seats" binding:"omitempty,min=1"`
	Price       float64 `json:"price" binding:"omitempty,min=0"`
	Status      string  `json:"status" binding:"omitempty,oneof=draft published sold_out cancelled completed"`
	ImageURL    string  `json:"image_url" binding:"omitempty,url"`
}

type EventResponse struct {
	ID             uuid.UUID   `json:"id"`
	Title          string      `json:"title"`
	Description    string      `json:"description"`
	Venue          string      `json:"venue"`
	EventDate      time.Time   `json:"event_date"`
	TotalSeats     int         `json:"total_seats"`
	AvailableSeats int         `json:"available_seats"`
	Price          float64     `json:"price"`
	Status         EventStatus `json:"status"`
	CreatedBy      uuid.UUID   `json:"created_by"`
	ImageURL       string      `json:"image_url,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}
