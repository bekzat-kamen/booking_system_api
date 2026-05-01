package handler

import (
	"net/http"
	"strconv"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/bekzat-kamen/booking_system_api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminEventHandler struct {
	adminEventService service.AdminEventServiceInterface
}

func NewAdminEventHandler(adminEventService service.AdminEventServiceInterface) *AdminEventHandler {
	return &AdminEventHandler{adminEventService: adminEventService}
}

// GetAllEvents godoc
// @Summary [RU] Список всех мероприятий (админ) / [EN] All events list (admin)
// @Description [RU] Расширенный список мероприятий с фильтрацией по статусу и организатору. / [EN] Extended events list with status and organizer filtering.
// @Tags admin-events
// @Produce  json
// @Param page query int false "[RU] Страница / [EN] Page"
// @Param limit query int false "[RU] Лимит / [EN] Limit"
// @Param status query string false "[RU] Статус / [EN] Status"
// @Param organizer_id query string false "[RU] ID организатора / [EN] Organizer ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /admin/events [get]
func (h *AdminEventHandler) GetAllEvents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	organizerID := c.Query("organizer_id")

	events, total, err := h.adminEventService.GetAllEvents(c.Request.Context(), page, limit, status, organizerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// GetEventDetail godoc
// @Summary [RU] Детальная информация о мероприятии / [EN] Event details
// @Description [RU] Возвращает полную информацию о мероприятии, включая статистику бронирований. / [EN] Returns full event information including booking stats.
// @Tags admin-events
// @Produce  json
// @Param id path string true "[RU] ID мероприятия / [EN] Event ID"
// @Success 200 {object} model.EventDetail
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /admin/events/{id} [get]
func (h *AdminEventHandler) GetEventDetail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	detail, err := h.adminEventService.GetEventDetail(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrEventNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get event"})
		return
	}

	c.JSON(http.StatusOK, detail)
}

// UpdateEvent godoc
// @Summary [RU] Обновить мероприятие (админ) / [EN] Update event (admin)
// @Description [RU] Принудительное обновление данных мероприятия администратором. / [EN] Forced update of event data by admin.
// @Tags admin-events
// @Accept  json
// @Produce  json
// @Param id path string true "[RU] ID мероприятия / [EN] Event ID"
// @Param event body model.UpdateEventRequest true "[RU] Данные для обновления / [EN] Update data"
// @Success 200 {object} model.Event
// @Security BearerAuth
// @Router /admin/events/{id} [put]
func (h *AdminEventHandler) UpdateEvent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	var req model.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := h.adminEventService.UpdateEvent(c.Request.Context(), id, &req)
	if err != nil {
		if err == repository.ErrEventNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update event"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// DeleteEvent godoc
// @Summary [RU] Отменить/Удалить мероприятие / [EN] Cancel/Delete event
// @Description [RU] Переводит мероприятие в статус CANCELLED. / [EN] Changes event status to CANCELLED.
// @Tags admin-events
// @Produce  json
// @Param id path string true "[RU] ID мероприятия / [EN] Event ID"
// @Success 200 {object} map[string]string
// @Security BearerAuth
// @Router /admin/events/{id} [delete]
func (h *AdminEventHandler) DeleteEvent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	err = h.adminEventService.DeleteEvent(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrEventNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "event cancelled successfully"})
}

// PublishEvent godoc
// @Summary [RU] Опубликовать мероприятие (админ) / [EN] Publish event (admin)
// @Description [RU] Опубликовать мероприятие от лица администратора. / [EN] Publish event as admin.
// @Tags admin-events
// @Produce  json
// @Param id path string true "[RU] ID мероприятия / [EN] Event ID"
// @Success 200 {object} model.Event
// @Security BearerAuth
// @Router /admin/events/{id}/publish [post]
func (h *AdminEventHandler) PublishEvent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	event, err := h.adminEventService.PublishEvent(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrEventNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish event"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// GetEventsStats godoc
// @Summary [RU] Статистика мероприятий / [EN] Events statistics
// @Description [RU] Общая статистика по всем мероприятиям в системе. / [EN] General statistics for all events in the system.
// @Tags admin-events
// @Produce  json
// @Success 200 {object} model.EventsStats
// @Security BearerAuth
// @Router /admin/events/stats [get]
func (h *AdminEventHandler) GetEventsStats(c *gin.Context) {
	stats, err := h.adminEventService.GetEventsStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
