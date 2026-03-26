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
	eventService *service.EventService
}

func NewEventHandler(eventService *service.EventService) *EventHandler {
	return &EventHandler{eventService: eventService}
}

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
