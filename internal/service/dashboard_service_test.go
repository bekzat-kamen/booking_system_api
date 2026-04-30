package service

import (
	"context"
	"errors"
	"testing"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type dashboardRepositoryMock struct {
	mock.Mock
}

func (m *dashboardRepositoryMock) GetStats(ctx context.Context) (*model.DashboardStats, error) {
	args := m.Called(ctx)
	stats, _ := args.Get(0).(*model.DashboardStats)
	return stats, args.Error(1)
}

func (m *dashboardRepositoryMock) GetRevenueByDay(ctx context.Context) ([]model.RevenueByDay, error) {
	args := m.Called(ctx)
	resp, _ := args.Get(0).([]model.RevenueByDay)
	return resp, args.Error(1)
}

func (m *dashboardRepositoryMock) GetTopEvents(ctx context.Context, limit int) ([]model.TopEvent, error) {
	args := m.Called(ctx, limit)
	resp, _ := args.Get(0).([]model.TopEvent)
	return resp, args.Error(1)
}

func (m *dashboardRepositoryMock) GetRecentActivities(ctx context.Context, limit int) ([]model.RecentActivity, error) {
	args := m.Called(ctx, limit)
	resp, _ := args.Get(0).([]model.RecentActivity)
	return resp, args.Error(1)
}

func TestDashboardServiceGetDashboardStatsSuccessWithFallbacks(t *testing.T) {
	ctx := context.Background()
	repo := new(dashboardRepositoryMock)
	svc := NewDashboardService(repo)

	repo.On("GetStats", ctx).Return(&model.DashboardStats{TotalUsers: 10}, nil).Once()
	repo.On("GetRevenueByDay", ctx).Return([]model.RevenueByDay{{Date: "2026-04-30", Amount: 100}}, nil).Once()
	repo.On("GetTopEvents", ctx, 5).Return([]model.TopEvent(nil), errors.New("fail")).Once()
	repo.On("GetRecentActivities", ctx, 10).Return([]model.RecentActivity(nil), errors.New("fail")).Once()

	resp, err := svc.GetDashboardStats(ctx)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.EqualValues(t, 10, resp.Stats.TotalUsers)
	assert.Len(t, resp.RevenueChart, 1)
	assert.Empty(t, resp.TopEvents)
	assert.Empty(t, resp.RecentActivities)
	repo.AssertExpectations(t)
}

func TestDashboardServiceGetDashboardStatsStatsError(t *testing.T) {
	ctx := context.Background()
	repo := new(dashboardRepositoryMock)
	svc := NewDashboardService(repo)

	repo.On("GetStats", ctx).Return((*model.DashboardStats)(nil), errors.New("stats fail")).Once()

	resp, err := svc.GetDashboardStats(ctx)

	require.Error(t, err)
	assert.Nil(t, resp)
	repo.AssertExpectations(t)
}
