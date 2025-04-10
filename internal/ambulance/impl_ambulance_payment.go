package ambulance

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wac-project/wac-api/internal/ambulance"
	"github.com/wac-project/wac-api/internal/db_service"
)

// AmbulancePaymentsAPI defines the interface for payment-related operations.
type AmbulancePaymentsAPI interface {
	CreatePayment(c *gin.Context)
	DeletePayment(c *gin.Context)
}

// implAmbulancePaymentAPI is the concrete implementation of AmbulancePaymentsAPI.
type implAmbulancePaymentAPI struct {
}

// NewAmbulancePaymentApi creates a new instance of AmbulancePaymentsAPI.
func NewAmbulancePaymentApi() AmbulancePaymentsAPI {
	return &implAmbulancePaymentAPI{}
}

// CreatePayment handles POST /api/payments.
// It creates a new payment record using the generic database service.
func (o implAmbulancePaymentAPI) CreatePayment(c *gin.Context) {
	value, exists := c.Get("db_service")
	if !exists {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db not found",
				"error":   "db not found",
			})
		return
	}

	db, ok := value.(db_service.DbService[ambulance.Payment])
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db context is not of required type",
				"error":   "cannot cast db context to db_service.DbService",
			})
		return
	}

	var payment ambulance.Payment
	if err := c.BindJSON(&payment); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":  "Bad Request",
				"message": "Invalid request body",
				"error":   err.Error(),
			})
		return
	}

	// Generate a new UUID if the payment ID is empty.
	if payment.Id == "" {
		payment.Id = uuid.New().String()
	}

	err := db.CreateDocument(c, payment.Id, &payment)
	switch err {
	case nil:
		c.JSON(http.StatusCreated, payment)
	case db_service.ErrConflict:
		c.JSON(
			http.StatusConflict,
			gin.H{
				"status":  "Conflict",
				"message": "Payment already exists",
				"error":   err.Error(),
			})
	default:
		c.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to create payment in database",
				"error":   err.Error(),
			})
	}
}

// DeletePayment handles DELETE /api/payments/:paymentId.
// It deletes the specified payment record from the database.
func (o implAmbulancePaymentAPI) DeletePayment(c *gin.Context) {
	value, exists := c.Get("db_service")
	if !exists {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db_service not found",
				"error":   "db_service not found",
			})
		return
	}

	db, ok := value.(db_service.DbService[ambulance.Payment])
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db_service context is not of type db_service.DbService",
				"error":   "cannot cast db_service context to db_service.DbService",
			})
		return
	}

	paymentId := c.Param("paymentId")
	err := db.DeleteDocument(c, paymentId)
	switch err {
	case nil:
		c.AbortWithStatus(http.StatusNoContent)
	case db_service.ErrNotFound:
		c.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Payment not found",
				"error":   err.Error(),
			})
	default:
		c.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to delete payment from database",
				"error":   err.Error(),
			})
	}
}
