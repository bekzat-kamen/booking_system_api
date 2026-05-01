package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/bekzat-kamen/booking_system_api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EventHandler struct {
	eventService service.EventServiceInterface
}

func NewEventHandler(eventService service.EventServiceInterface) *EventHandler {
	return &EventHandler{eventService: eventService}
}

// Create godoc
// @Summary [RU] Создать мероприятие / [EN] Create event
// @Description [RU] Создает новое мероприятие. Требуется авторизация. / [EN] Creates a new event. Authorization required.
// @Tags events
// @Accept  json
// @Produce  json
// @Param event body model.CreateEventRequest true "[RU] Данные мероприятия / [EN] Event data"
// @Success 201 {object} model.Event
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /events [post]
func (h *EventHandler) Create(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	var req model.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := h.eventService.Create(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidEventDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "event_date must be in RFC3339 format"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, event)
}

// GetByID godoc
// @Summary [RU] Получить мероприятие по ID / [EN] Get event by ID
// @Description [RU] Возвращает полную информацию о мероприятии по его идентификатору. / [EN] Returns full information about an event by its ID.
// @Tags events
// @Produce  json
// @Param id path string true "[RU] ID мероприятия / [EN] Event ID"
// @Success 200 {object} model.Event
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /events/{id} [get]
func (h *EventHandler) GetByID(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	event, err := h.eventService.GetByID(c.Request.Context(), eventID)
	if err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get event"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// GetAll godoc
// @Summary [RU] Получить все мероприятия / [EN] Get all events
// @Description [RU] Возвращает список всех мероприятий с поддержкой пагинации. / [EN] Returns a list of all events with pagination support.
// @Tags events
// @Produce  json
// @Param limit query int false "[RU] Лимит (по умолчанию 20) / [EN] Limit (default 20)"
// @Param page query int false "[RU] Страница (по умолчанию 1) / [EN] Page (default 1)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /events [get]
func (h *EventHandler) GetAll(c *gin.Context) {
	limit, page := parsePagination(c)
	offset := (page - 1) * limit

	events, total, err := h.eventService.GetAll(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": events,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"page":   page,
			"offset": offset,
		},
	})
}

// GetByOrganizer godoc
// @Summary [RU] Получить мероприятия организатора / [EN] Get organizer events
// @Description [RU] Возвращает список мероприятий, созданных текущим пользователем. / [EN] Returns a list of events created by the current user.
// @Tags events
// @Produce  json
// @Param limit query int false "[RU] Лимит / [EN] Limit"
// @Param page query int false "[RU] Страница / [EN] Page"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /events/organizer [get]
func (h *EventHandler) GetByOrganizer(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	limit, page := parsePagination(c)
	offset := (page - 1) * limit

	events, err := h.eventService.GetByOrganizer(c.Request.Context(), userID.(uuid.UUID), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get organizer events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": events,
		"meta": gin.H{
			"limit":  limit,
			"page":   page,
			"offset": offset,
		},
	})
}

// Update godoc
// @Summary [RU] Обновить мероприятие / [EN] Update event
// @Description [RU] Обновляет данные существующего мероприятия. Только для организатора. / [EN] Updates existing event data. For organizer only.
// @Tags events
// @Accept  json
// @Produce  json
// @Param id path string true "[RU] ID мероприятия / [EN] Event ID"
// @Param event body model.UpdateEventRequest true "[RU] Данные для обновления / [EN] Update data"
// @Success 200 {object} model.Event
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /events/{id} [put]
func (h *EventHandler) Update(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	var req model.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := h.eventService.Update(c.Request.Context(), eventID, userID.(uuid.UUID), &req)
	if err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		if errors.Is(err, service.ErrEventForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "you are not allowed to modify this event"})
			return
		}
		if errors.Is(err, service.ErrInvalidEventDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "event_date must be in RFC3339 format"})
			return
		}
		if errors.Is(err, service.ErrInvalidSeatCount) || errors.Is(err, service.ErrInvalidSeatBalance) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update event"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// Delete godoc
// @Summary [RU] Удалить мероприятие / [EN] Delete event
// @Description [RU] Удаляет мероприятие из системы. Только для организатора. / [EN] Deletes an event from the system. For organizer only.
// @Tags events
// @Produce  json
// @Param id path string true "[RU] ID мероприятия / [EN] Event ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /events/{id} [delete]
func (h *EventHandler) Delete(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	err = h.eventService.Delete(c.Request.Context(), eventID, userID.(uuid.UUID))
	if err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		if errors.Is(err, service.ErrEventForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "you are not allowed to delete this event"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "event deleted successfully"})
}

// PublishEvent godoc
// @Summary [RU] Опубликовать мероприятие / [EN] Publish event
// @Description [RU] Переводит статус мероприятия в "опубликовано". / [EN] Changes event status to "published".
// @Tags events
// @Produce  json
// @Param id path string true "[RU] ID мероприятия / [EN] Event ID"
// @Success 200 {object} model.Event
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /events/{id}/publish [post]
func (h *EventHandler) PublishEvent(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	event, err := h.eventService.PublishEvent(c.Request.Context(), eventID, userID.(uuid.UUID))
	if err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		if errors.Is(err, service.ErrEventForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "you are not allowed to publish this event"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish event"})
		return
	}

	c.JSON(http.StatusOK, event)
}

func parsePagination(c *gin.Context) (int, int) {
	limit := 20
	page := 1

	if limitStr := c.Query("limit"); limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if v, err := strconv.Atoi(pageStr); err == nil && v > 0 {
			page = v
		}
	}

	return limit, page
}
