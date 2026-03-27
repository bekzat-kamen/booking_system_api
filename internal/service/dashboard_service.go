package service

import (
	"context"
	"time"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
)

type DashboardService struct {
	dashboardRepo *repository.DashboardRepository
}

func NewDashboardService(dashboardRepo *repository.DashboardRepository) *DashboardService {
	return &DashboardService{dashboardRepo: dashboardRepo}
}

func (s *DashboardService) GetDashboardStats(ctx context.Context) (*model.DashboardResponse, error) {

	stats, err := s.dashboardRepo.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	revenueChart, err := s.dashboardRepo.GetRevenueByDay(ctx)
	if err != nil {
		revenueChart = []model.RevenueByDay{}
	}

	topEvents, err := s.dashboardRepo.GetTopEvents(ctx, 5)
	if err != nil {
		topEvents = []model.TopEvent{}
	}

	recentActivities, err := s.dashboardRepo.GetRecentActivities(ctx, 10)
	if err != nil {
		recentActivities = []model.RecentActivity{}
	}

	return &model.DashboardResponse{
		Stats:            stats,
		RevenueChart:     revenueChart,
		TopEvents:        topEvents,
		RecentActivities: recentActivities,
		GeneratedAt:      time.Now(),
	}, nil
}
