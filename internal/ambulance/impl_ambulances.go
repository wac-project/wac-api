package ambulance

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
	"github.com/wac-project/wac-api/internal/db_service"
)

// Ambulance represents the structure for an ambulance.
// Ensure that this matches your overall project data model.
type Ambulance struct {
    Id         string `json:"id" bson:"_id,omitempty"`
    // Add any additional fields as required.
}

// AmbulancesAPI defines the interface for ambulance related operations.
type AmbulancesAPI interface {
    CreateAmbulance(c *gin.Context)
    DeleteAmbulance(c *gin.Context)
}

// implAmbulancesAPI is the concrete implementation of AmbulancesAPI.
type implAmbulancesAPI struct {
}

// NewAmbulancesApi creates a new AmbulancesAPI instance.
func NewAmbulancesApi() AmbulancesAPI {
    return &implAmbulancesAPI{}
}

// CreateAmbulance handles POST /api/ambulances.
// It reads the ambulance data from the request, assigns a new UUID if needed,
// and creates the document using the db_service.
func (o implAmbulancesAPI) CreateAmbulance(c *gin.Context) {
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

    db, ok := value.(db_service.DbService[Ambulance])
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

    ambulance := Ambulance{}
    err := c.BindJSON(&ambulance)
    if err != nil {
        c.JSON(
            http.StatusBadRequest,
            gin.H{
                "status":  "Bad Request",
                "message": "Invalid request body",
                "error":   err.Error(),
            })
        return
    }

    // Generate a new UUID if the ID is empty.
    if ambulance.Id == "" {
        ambulance.Id = uuid.New().String()
    }

    err = db.CreateDocument(c, ambulance.Id, &ambulance)
    switch err {
    case nil:
        c.JSON(
            http.StatusCreated,
            ambulance,
        )
    case db_service.ErrConflict:
        c.JSON(
            http.StatusConflict,
            gin.H{
                "status":  "Conflict",
                "message": "Ambulance already exists",
                "error":   err.Error(),
            },
        )
    default:
        c.JSON(
            http.StatusBadGateway,
            gin.H{
                "status":  "Bad Gateway",
                "message": "Failed to create ambulance in database",
                "error":   err.Error(),
            },
        )
    }
}

// DeleteAmbulance handles DELETE /api/ambulances/:ambulanceId.
// It deletes the specified ambulance document using the db_service.
func (o implAmbulancesAPI) DeleteAmbulance(c *gin.Context) {
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

    db, ok := value.(db_service.DbService[Ambulance])
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

    ambulanceId := c.Param("ambulanceId")
    err := db.DeleteDocument(c, ambulanceId)

    switch err {
    case nil:
        c.AbortWithStatus(http.StatusNoContent)
    case db_service.ErrNotFound:
        c.JSON(
            http.StatusNotFound,
            gin.H{
                "status":  "Not Found",
                "message": "Ambulance not found",
                "error":   err.Error(),
            },
        )
    default:
        c.JSON(
            http.StatusBadGateway,
            gin.H{
                "status":  "Bad Gateway",
                "message": "Failed to delete ambulance from database",
                "error":   err.Error(),
            },
        )
    }
}
