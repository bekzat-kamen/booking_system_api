package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type dashboardServiceMock struct {
	mock.Mock
}

func (m *dashboardServiceMock) GetDashboardStats(ctx context.Context) (*model.DashboardResponse, error) {
	args := m.Called(ctx)
	resp, _ := args.Get(0).(*model.DashboardResponse)
	return resp, args.Error(1)
}

func TestDashboardHandlerGetStatsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(dashboardServiceMock)
	h := NewDashboardHandler(svc)
	router := gin.New()
	router.GET("/dashboard", h.GetStats)

	svc.On("GetDashboardStats", mock.Anything).Return(&model.DashboardResponse{}, nil).Once()

	w := performJSONRequest(router, http.MethodGet, "/dashboard", nil)

	require.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestDashboardHandlerGetStatsError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(dashboardServiceMock)
	h := NewDashboardHandler(svc)
	router := gin.New()
	router.GET("/dashboard", h.GetStats)

	svc.On("GetDashboardStats", mock.Anything).Return((*model.DashboardResponse)(nil), errors.New("fail")).Once()

	w := performJSONRequest(router, http.MethodGet, "/dashboard", nil)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to get dashboard stats")
}
