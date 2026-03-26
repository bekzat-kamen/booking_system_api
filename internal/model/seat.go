package model

import (
	"time"

	"github.com/google/uuid"
)

type SeatStatus string

const (
	SeatStatusAvailable SeatStatus = "available"
	SeatStatusBooked    SeatStatus = "booked"
	SeatStatusReserved  SeatStatus = "reserved"
	SeatStatusBlocked   SeatStatus = "blocked"
)

type Seat struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	EventID    uuid.UUID  `db:"event_id" json:"event_id"`
	SeatNumber string     `db:"seat_number" json:"seat_number"`
	RowNumber  string     `db:"row_number" json:"row_number"`
	Status     SeatStatus `db:"status" json:"status"`
	Price      float64    `db:"price" json:"price"`
	Version    int        `db:"version" json:"version"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`
}

type CreateSeatRequest struct {
	EventID    uuid.UUID
	SeatNumber string
	RowNumber  string
	Price      float64
}

type GenerateSeatsRequest struct {
	TotalRows   int     `json:"total_rows" binding:"required,min=1,max=50"`
	SeatsPerRow int     `json:"seats_per_row" binding:"required,min=1,max=100"`
	BasePrice   float64 `json:"base_price" binding:"required,min=0"`
}

type SeatResponse struct {
	ID         uuid.UUID  `json:"id"`
	EventID    uuid.UUID  `json:"event_id"`
	SeatNumber string     `json:"seat_number"`
	RowNumber  string     `json:"row_number"`
	Status     SeatStatus `json:"status"`
	Price      float64    `json:"price"`
}

type SeatMapResponse struct {
	EventID    uuid.UUID      `json:"event_id"`
	TotalSeats int            `json:"total_seats"`
	Available  int            `json:"available"`
	Booked     int            `json:"booked"`
	Reserved   int            `json:"reserved"`
	Blocked    int            `json:"blocked"`
	Seats      []SeatResponse `json:"seats"`
}
