package repository

import (
	"context"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/jmoiron/sqlx"
)

type DashboardRepositoryInterface interface {
	GetStats(ctx context.Context) (*model.DashboardStats, error)
	GetRevenueByDay(ctx context.Context) ([]model.RevenueByDay, error)
	GetTopEvents(ctx context.Context, limit int) ([]model.TopEvent, error)
	GetRecentActivities(ctx context.Context, limit int) ([]model.RecentActivity, error)
}

type DashboardRepository struct {
	db *sqlx.DB
}

func NewDashboardRepository(db *sqlx.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) GetStats(ctx context.Context) (*model.DashboardStats, error) {
	stats := &model.DashboardStats{}

	if err := r.db.GetContext(ctx, &stats.TotalUsers, `SELECT COUNT(*) FROM users`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &stats.NewUsersToday, `SELECT COUNT(*) FROM users WHERE created_at >= NOW() - INTERVAL '1 day'`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &stats.ActiveUsers, `SELECT COUNT(*) FROM users WHERE status = 'active'`); err != nil {
		return nil, err
	}

	if err := r.db.GetContext(ctx, &stats.TotalEvents, `SELECT COUNT(*) FROM events`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &stats.PublishedEvents, `SELECT COUNT(*) FROM events WHERE status = 'published'`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &stats.DraftEvents, `SELECT COUNT(*) FROM events WHERE status = 'draft'`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &stats.CancelledEvents, `SELECT COUNT(*) FROM events WHERE status = 'cancelled'`); err != nil {
		return nil, err
	}

	if err := r.db.GetContext(ctx, &stats.TotalBookings, `SELECT COUNT(*) FROM bookings`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &stats.PendingBookings, `SELECT COUNT(*) FROM bookings WHERE status = 'pending'`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &stats.ConfirmedBookings, `SELECT COUNT(*) FROM bookings WHERE status = 'confirmed'`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &stats.CancelledBookings, `SELECT COUNT(*) FROM bookings WHERE status = 'cancelled'`); err != nil {
		return nil, err
	}

	if err := r.db.GetContext(ctx, &stats.TotalRevenue, `SELECT COALESCE(SUM(final_amount), 0) FROM bookings WHERE status = 'confirmed'`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &stats.TodayRevenue, `SELECT COALESCE(SUM(final_amount), 0) FROM bookings WHERE status = 'confirmed' AND paid_at >= NOW() - INTERVAL '1 day'`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &stats.MonthRevenue, `SELECT COALESCE(SUM(final_amount), 0) FROM bookings WHERE status = 'confirmed' AND paid_at >= NOW() - INTERVAL '30 day'`); err != nil {
		return nil, err
	}

	if err := r.db.GetContext(ctx, &stats.AverageCheck, `SELECT COALESCE(AVG(final_amount), 0) FROM bookings WHERE status = 'confirmed'`); err != nil {
		return nil, err
	}

	if err := r.db.GetContext(ctx, &stats.TotalPromocodes, `SELECT COUNT(*) FROM promocodes`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &stats.ActivePromocodes, `SELECT COUNT(*) FROM promocodes WHERE is_active = true`); err != nil {
		return nil, err
	}

	if err := r.db.GetContext(ctx, &stats.TotalPayments, `SELECT COUNT(*) FROM payment_transactions`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &stats.SuccessPayments, `SELECT COUNT(*) FROM payment_transactions WHERE status = 'success'`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &stats.FailedPayments, `SELECT COUNT(*) FROM payment_transactions WHERE status = 'failed'`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &stats.RefundedPayments, `SELECT COUNT(*) FROM payment_transactions WHERE status = 'refunded'`); err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *DashboardRepository) GetRevenueByDay(ctx context.Context) ([]model.RevenueByDay, error) {
	query := `
		SELECT 
			TO_CHAR(paid_at, 'YYYY-MM-DD') as date,
			COALESCE(SUM(final_amount), 0) as amount
		FROM bookings
		WHERE status = 'confirmed' 
		  AND paid_at >= NOW() - INTERVAL '30 day'
		GROUP BY TO_CHAR(paid_at, 'YYYY-MM-DD')
		ORDER BY date ASC
	`

	var result []model.RevenueByDay
	err := r.db.SelectContext(ctx, &result, query)
	return result, err
}

func (r *DashboardRepository) GetTopEvents(ctx context.Context, limit int) ([]model.TopEvent, error) {
	query := `
		SELECT 
			e.id as event_id,
			e.title,
			COUNT(b.id) as bookings,
			COALESCE(SUM(b.final_amount), 0) as revenue,
			e.event_date
		FROM events e
		LEFT JOIN bookings b ON e.id = b.event_id AND b.status = 'confirmed'
		GROUP BY e.id, e.title, e.event_date
		ORDER BY revenue DESC
		LIMIT $1
	`

	var result []model.TopEvent
	err := r.db.SelectContext(ctx, &result, query, limit)
	return result, err
}

func (r *DashboardRepository) GetRecentActivities(ctx context.Context, limit int) ([]model.RecentActivity, error) {
	query := `
		SELECT 
			al.id,
			al.action,
			al.entity_type as entity,
			u.email as user_email,
			al.created_at
		FROM audit_logs al
		LEFT JOIN users u ON al.user_id = u.id
		ORDER BY al.created_at DESC
		LIMIT $1
	`

	var result []model.RecentActivity
	err := r.db.SelectContext(ctx, &result, query, limit)
	return result, err
}
