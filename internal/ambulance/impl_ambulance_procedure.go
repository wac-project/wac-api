package ambulance

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/SimonK1/ambulance-webapi/internal/db_service"
	"github.com/SimonK1/ambulance-webapi/ambulance"
)

// AmbulanceProceduresAPI defines the interface for procedure-related operations.
type AmbulanceProceduresAPI interface {
	CreateProcedure(c *gin.Context)
	DeleteProcedure(c *gin.Context)
}

// implAmbulanceProcedureAPI is the concrete implementation of AmbulanceProceduresAPI.
type implAmbulanceProcedureAPI struct {
}

// NewAmbulanceProcedureApi creates a new instance of AmbulanceProceduresAPI.
func NewAmbulanceProcedureApi() AmbulanceProceduresAPI {
	return &implAmbulanceProcedureAPI{}
}

// CreateProcedure handles POST /api/procedures.
// It creates a new procedure record using the generic database service.
func (o implAmbulanceProcedureAPI) CreateProcedure(c *gin.Context) {
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

	db, ok := value.(db_service.DbService[ambulance.Procedure])
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

	var procedure ambulance.Procedure
	if err := c.BindJSON(&procedure); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":  "Bad Request",
				"message": "Invalid request body",
				"error":   err.Error(),
			})
		return
	}

	// Generate a new UUID if the procedure ID is empty.
	if procedure.Id == "" {
		procedure.Id = uuid.New().String()
	}

	err := db.CreateDocument(c, procedure.Id, &procedure)
	switch err {
	case nil:
		c.JSON(http.StatusCreated, procedure)
	case db_service.ErrConflict:
		c.JSON(
			http.StatusConflict,
			gin.H{
				"status":  "Conflict",
				"message": "Procedure already exists",
				"error":   err.Error(),
			})
	default:
		c.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to create procedure in database",
				"error":   err.Error(),
			})
	}
}

// DeleteProcedure handles DELETE /api/procedures/:procedureId.
// It deletes the specified procedure record from the database.
func (o implAmbulanceProcedureAPI) DeleteProcedure(c *gin.Context) {
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

	db, ok := value.(db_service.DbService[ambulance.Procedure])
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

	procedureId := c.Param("procedureId")
	err := db.DeleteDocument(c, procedureId)
	switch err {
	case nil:
		c.AbortWithStatus(http.StatusNoContent)
	case db_service.ErrNotFound:
		c.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Procedure not found",
				"error":   err.Error(),
			})
		return
	default:
		c.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to delete procedure from database",
				"error":   err.Error(),
			})
	}
}
