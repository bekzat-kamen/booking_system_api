package handler

import (
	"net/http"

	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/bekzat-kamen/booking_system_api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
}

func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req model.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.paymentService.CreatePayment(c.Request.Context(), &req)
	if err != nil {
		if err == repository.ErrBookingNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		if err == service.ErrPaymentAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payment"})
		return
	}

	c.JSON(http.StatusCreated, payment)
}

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
		return
	}

	payment, err := h.paymentService.ProcessPayment(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrPaymentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		if err == service.ErrInvalidPaymentStatus {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process payment"})
		return
	}

	c.JSON(http.StatusOK, payment)
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
		return
	}

	payment, err := h.paymentService.GetPayment(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrPaymentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get payment"})
		return
	}

	c.JSON(http.StatusOK, payment)
}

func (h *PaymentHandler) Webhook(c *gin.Context) {
	var req model.WebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.paymentService.ProcessWebhook(c.Request.Context(), &req)
	if err != nil {
		if err == repository.ErrPaymentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "webhook processed successfully"})
}
