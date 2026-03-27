package repository

import (
	"context"
	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/jmoiron/sqlx"
)

type DashboardRepository struct {
	db *sqlx.DB
}

func NewDashboardRepository(db *sqlx.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) GetStats(ctx context.Context) (*model.DashboardStats, error) {
	stats := &model.DashboardStats{}

	r.db.GetContext(ctx, &stats.TotalUsers, `SELECT COUNT(*) FROM users`)
	r.db.GetContext(ctx, &stats.NewUsersToday, `SELECT COUNT(*) FROM users WHERE created_at >= NOW() - INTERVAL '1 day'`)
	r.db.GetContext(ctx, &stats.ActiveUsers, `SELECT COUNT(*) FROM users WHERE status = 'active'`)

	r.db.GetContext(ctx, &stats.TotalEvents, `SELECT COUNT(*) FROM events`)
	r.db.GetContext(ctx, &stats.PublishedEvents, `SELECT COUNT(*) FROM events WHERE status = 'published'`)
	r.db.GetContext(ctx, &stats.DraftEvents, `SELECT COUNT(*) FROM events WHERE status = 'draft'`)
	r.db.GetContext(ctx, &stats.CancelledEvents, `SELECT COUNT(*) FROM events WHERE status = 'cancelled'`)

	r.db.GetContext(ctx, &stats.TotalBookings, `SELECT COUNT(*) FROM bookings`)
	r.db.GetContext(ctx, &stats.PendingBookings, `SELECT COUNT(*) FROM bookings WHERE status = 'pending'`)
	r.db.GetContext(ctx, &stats.ConfirmedBookings, `SELECT COUNT(*) FROM bookings WHERE status = 'confirmed'`)
	r.db.GetContext(ctx, &stats.CancelledBookings, `SELECT COUNT(*) FROM bookings WHERE status = 'cancelled'`)

	r.db.GetContext(ctx, &stats.TotalRevenue, `SELECT COALESCE(SUM(final_amount), 0) FROM bookings WHERE status = 'confirmed'`)
	r.db.GetContext(ctx, &stats.TodayRevenue, `SELECT COALESCE(SUM(final_amount), 0) FROM bookings WHERE status = 'confirmed' AND paid_at >= NOW() - INTERVAL '1 day'`)
	r.db.GetContext(ctx, &stats.MonthRevenue, `SELECT COALESCE(SUM(final_amount), 0) FROM bookings WHERE status = 'confirmed' AND paid_at >= NOW() - INTERVAL '30 day'`)

	r.db.GetContext(ctx, &stats.AverageCheck, `SELECT COALESCE(AVG(final_amount), 0) FROM bookings WHERE status = 'confirmed'`)

	r.db.GetContext(ctx, &stats.TotalPromocodes, `SELECT COUNT(*) FROM promocodes`)
	r.db.GetContext(ctx, &stats.ActivePromocodes, `SELECT COUNT(*) FROM promocodes WHERE is_active = true`)

	r.db.GetContext(ctx, &stats.TotalPayments, `SELECT COUNT(*) FROM payment_transactions`)
	r.db.GetContext(ctx, &stats.SuccessPayments, `SELECT COUNT(*) FROM payment_transactions WHERE status = 'success'`)
	r.db.GetContext(ctx, &stats.FailedPayments, `SELECT COUNT(*) FROM payment_transactions WHERE status = 'failed'`)
	r.db.GetContext(ctx, &stats.RefundedPayments, `SELECT COUNT(*) FROM payment_transactions WHERE status = 'refunded'`)

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
	if err != nil {

		return []model.RecentActivity{}, nil
	}

	return result, err
}
