package model

// UserDetail представляет расширенную информацию о пользователе для админки
type UserDetail struct {
	User       User                   `json:"user"`
	Statistics map[string]interface{} `json:"statistics"`
}

// UserStats представляет общую статистику по пользователям
type UserStats struct {
	TotalUsers      int64 `json:"total_users"`
	ActiveUsers     int64 `json:"active_users"`
	BlockedUsers    int64 `json:"blocked_users"`
	UnverifiedUsers int64 `json:"unverified_users"`
}

// EventDetail представляет расширенную информацию о мероприятии для админки
type EventDetail struct {
	Event      Event                  `json:"event"`
	Statistics map[string]interface{} `json:"statistics"`
}

// EventsStats представляет общую статистику по мероприятиям
type EventsStats struct {
	TotalEvents     int64 `json:"total_events"`
	PublishedEvents int64 `json:"published_events"`
	DraftEvents     int64 `json:"draft_events"`
	CancelledEvents int64 `json:"cancelled_events"`
}

// BookingDetail представляет расширенную информацию о бронировании для админки
type BookingDetail struct {
	Booking Booking        `json:"booking"`
	Seats   []*BookingSeat `json:"seats"`
	Payment *Payment       `json:"payment"`
}

// BookingsStats представляет общую статистику по бронированиям
type BookingsStats struct {
	TotalBookings     int64   `json:"total_bookings"`
	PendingBookings   int64   `json:"pending_bookings"`
	ConfirmedBookings int64   `json:"confirmed_bookings"`
	CancelledBookings int64   `json:"cancelled_bookings"`
	TotalRevenue      float64 `json:"total_revenue"`
}

// PromocodeDetail представляет расширенную информацию о промокоде для админки
type PromocodeDetail struct {
	Promocode Promocode              `json:"promocode"`
	UsageLogs map[string]interface{} `json:"usage_logs"`
}

// PromocodesStats представляет общую статистику по промокодам
type PromocodesStats struct {
	TotalPromocodes  int64 `json:"total_promocodes"`
	ActivePromocodes int64 `json:"active_promocodes"`
	TotalUses        int64 `json:"total_uses"`
}
