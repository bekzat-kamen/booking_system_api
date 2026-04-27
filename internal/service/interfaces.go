package service

import (
	"context"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/google/uuid"
)

type TokenService interface {
	GenerateTokens(userID uuid.UUID, email string, role string) (string, string, error)
	ValidateAccessToken(tokenString string) (*Claims, error)
	ValidateRefreshToken(tokenString string) (*Claims, error)
	RefreshTokens(refreshToken string) (string, string, error)
}

type TokenValidator interface {
	ValidateAccessToken(tokenString string) (*Claims, error)
}

type AuthServiceInterface interface {
	Register(ctx context.Context, req *model.CreateUserRequest) (*model.User, error)
	Login(ctx context.Context, req *model.LoginRequest) (*LoginResponse, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req *model.UpdateUserRequest) (*model.User, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, req *model.ChangePasswordRequest) error
	DeactivateProfile(ctx context.Context, userID uuid.UUID) error
	ValidateEmail(ctx context.Context, email string) (bool, error)
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
}

type EventServiceInterface interface {
	Create(ctx context.Context, organizerID uuid.UUID, req *model.CreateEventRequest) (*model.Event, error)
	GetByID(ctx context.Context, eventID uuid.UUID) (*model.Event, error)
	GetAll(ctx context.Context, limit, offset int) ([]*model.Event, int, error)
	GetByOrganizer(ctx context.Context, organizerID uuid.UUID, limit, offset int) ([]*model.Event, error)
	Update(ctx context.Context, eventID, organizerID uuid.UUID, req *model.UpdateEventRequest) (*model.Event, error)
	Delete(ctx context.Context, eventID, organizerID uuid.UUID) error
	PublishEvent(ctx context.Context, eventID, organizerID uuid.UUID) (*model.Event, error)
}

type SeatServiceInterface interface {
	GenerateSeats(ctx context.Context, eventID, organizerID uuid.UUID, req *model.GenerateSeatsRequest) error
	GetSeatMap(ctx context.Context, eventID uuid.UUID) (*model.SeatMapResponse, error)
	GetAvailableSeats(ctx context.Context, eventID uuid.UUID) ([]model.SeatResponse, error)
	ReserveSeat(ctx context.Context, seatID uuid.UUID) error
	BookSeat(ctx context.Context, seatID uuid.UUID) error
	ReleaseSeat(ctx context.Context, seatID uuid.UUID) error
}

type BookingServiceInterface interface {
	CreateBooking(ctx context.Context, userID uuid.UUID, req *model.CreateBookingRequest) (*model.BookingResponse, error)
	GetBooking(ctx context.Context, id uuid.UUID) (*model.BookingResponse, error)
	GetUserBookings(ctx context.Context, userID uuid.UUID, page, limit int) ([]*model.BookingResponse, int, error)
	CancelBooking(ctx context.Context, bookingID uuid.UUID, userID uuid.UUID) error
	ConfirmBooking(ctx context.Context, bookingID uuid.UUID) error
	ExpirePendingBookings(ctx context.Context) (int, error)
}

type PaymentServiceInterface interface {
	CreatePayment(ctx context.Context, req *model.CreatePaymentRequest) (*model.Payment, error)
	ProcessPayment(ctx context.Context, paymentID uuid.UUID) (*model.Payment, error)
	ProcessWebhook(ctx context.Context, req *model.WebhookRequest) error
	GetPayment(ctx context.Context, id uuid.UUID) (*model.PaymentResponse, error)
	GetPaymentByBooking(ctx context.Context, bookingID uuid.UUID) (*model.PaymentResponse, error)
}

type PromocodeServiceInterface interface {
	CreatePromocode(ctx context.Context, createdBy uuid.UUID, req *model.CreatePromocodeRequest) (*model.Promocode, error)
	ValidatePromocode(ctx context.Context, req *model.ValidatePromocodeRequest) (*model.PromocodeValidationResponse, error)
	ApplyPromocode(ctx context.Context, code string) error
	GetPromocode(ctx context.Context, id uuid.UUID) (*model.Promocode, error)
	GetAllPromocodes(ctx context.Context, page, limit int) ([]*model.Promocode, int, error)
	UpdatePromocode(ctx context.Context, id uuid.UUID, req *model.UpdatePromocodeRequest) (*model.Promocode, error)
	DeletePromocode(ctx context.Context, id uuid.UUID) error
	DeactivatePromocode(ctx context.Context, id uuid.UUID) error
}

type DashboardServiceInterface interface {
	GetDashboardStats(ctx context.Context) (*model.DashboardResponse, error)
}

type AdminUserServiceInterface interface {
	GetAllUsers(ctx context.Context, page, limit int, status, role string) ([]*model.User, int, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetUserDetail(ctx context.Context, id uuid.UUID) (map[string]interface{}, error)
	UpdateUserRole(ctx context.Context, id uuid.UUID, role model.Role, adminID uuid.UUID) error
	BlockUser(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error
	UnblockUser(ctx context.Context, id uuid.UUID) error
	GetUserStats(ctx context.Context) (map[string]int64, error)
	DeleteUser(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error
}

type AdminEventServiceInterface interface {
	GetAllEvents(ctx context.Context, page, limit int, status, organizerID string) ([]*model.Event, int, error)
	GetEventDetail(ctx context.Context, id uuid.UUID) (map[string]interface{}, error)
	UpdateEvent(ctx context.Context, id uuid.UUID, req *model.UpdateEventRequest) (*model.Event, error)
	DeleteEvent(ctx context.Context, id uuid.UUID) error
	PublishEvent(ctx context.Context, id uuid.UUID) (*model.Event, error)
	GetEventsStats(ctx context.Context) (map[string]int64, error)
}

type AdminBookingServiceInterface interface {
	GetAllBookings(ctx context.Context, page, limit int, status, userID, eventID string) ([]*model.Booking, int, error)
	GetBookingDetail(ctx context.Context, id uuid.UUID) (map[string]interface{}, error)
	CancelBooking(ctx context.Context, id uuid.UUID, refund bool) error
	RefundBooking(ctx context.Context, id uuid.UUID) error
	GetBookingsStats(ctx context.Context) (map[string]interface{}, error)
	ExportBookingsToCSV(ctx context.Context, status string) ([][]string, error)
}

type AdminPromocodeServiceInterface interface {
	GetAllPromocodes(ctx context.Context, page, limit int, isActive string) ([]*model.Promocode, int, error)
	GetPromocodeDetail(ctx context.Context, id uuid.UUID) (map[string]interface{}, error)
	UpdatePromocode(ctx context.Context, id uuid.UUID, req *model.UpdatePromocodeRequest) (*model.Promocode, error)
	DeletePromocode(ctx context.Context, id uuid.UUID) error
	BulkDeactivate(ctx context.Context, ids []uuid.UUID) error
	GetPromocodesStats(ctx context.Context) (map[string]int64, error)
	ExportPromocodesToCSV(ctx context.Context) ([][]string, error)
}
