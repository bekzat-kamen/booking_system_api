package model

import (
	"time"

	"github.com/google/uuid"
)

type DashboardStats struct {
	TotalUsers    int64 `json:"total_users"`
	NewUsersToday int64 `json:"new_users_today"`
	ActiveUsers   int64 `json:"active_users"`

	TotalEvents     int64 `json:"total_events"`
	PublishedEvents int64 `json:"published_events"`
	DraftEvents     int64 `json:"draft_events"`
	CancelledEvents int64 `json:"cancelled_events"`

	TotalBookings     int64 `json:"total_bookings"`
	PendingBookings   int64 `json:"pending_bookings"`
	ConfirmedBookings int64 `json:"confirmed_bookings"`
	CancelledBookings int64 `json:"cancelled_bookings"`

	TotalRevenue float64 `json:"total_revenue"`
	TodayRevenue float64 `json:"today_revenue"`
	MonthRevenue float64 `json:"month_revenue"`
	AverageCheck float64 `json:"average_check"`

	TotalPromocodes  int64 `json:"total_promocodes"`
	ActivePromocodes int64 `json:"active_promocodes"`

	TotalPayments    int64 `json:"total_payments"`
	SuccessPayments  int64 `json:"success_payments"`
	FailedPayments   int64 `json:"failed_payments"`
	RefundedPayments int64 `json:"refunded_payments"`
}

type RevenueByDay struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
}

type TopEvent struct {
	EventID   uuid.UUID `json:"event_id"`
	Title     string    `json:"title"`
	Bookings  int64     `json:"bookings"`
	Revenue   float64   `json:"revenue"`
	EventDate time.Time `json:"event_date"`
}

type RecentActivity struct {
	ID        uuid.UUID `json:"id"`
	Action    string    `json:"action"`
	Entity    string    `json:"entity"`
	UserEmail string    `json:"user_email"`
	CreatedAt time.Time `json:"created_at"`
}

type DashboardResponse struct {
	Stats            *DashboardStats  `json:"stats"`
	RevenueChart     []RevenueByDay   `json:"revenue_chart"`
	TopEvents        []TopEvent       `json:"top_events"`
	RecentActivities []RecentActivity `json:"recent_activities"`
	GeneratedAt      time.Time        `json:"generated_at"`
}
