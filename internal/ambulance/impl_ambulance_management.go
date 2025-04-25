package ambulance

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// implAmbulanceAPI implements the AmbulanceManagementAPI interface.
type implAmbulanceAPI struct {
	collection *mongo.Collection
}

// NewAmbulanceAPI creates a new instance of AmbulanceManagementAPI using the given MongoDB collection.
func NewAmbulanceAPI(col *mongo.Collection) AmbulanceManagementAPI {
	return &implAmbulanceAPI{collection: col}
}

// updateAmbulanceFunc is a helper function that retrieves an ambulance by its ID from MongoDB,
// then passes it to the provided closure for further processing. If the ambulance is updated,
// it will be saved back to the database.
func updateAmbulanceFunc(c *gin.Context, fn func(c *gin.Context, ambulance *Ambulance) (*Ambulance, interface{}, int), col *mongo.Collection) {
	ambulanceId := c.Param("ambulanceId")
	if ambulanceId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Ambulance ID is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var ambulance Ambulance
	if err := col.FindOne(ctx, bson.M{"_id": ambulanceId}).Decode(&ambulance); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  http.StatusNotFound,
			"message": "Ambulance not found",
		})
		return
	}

	updatedAmbulance, result, statusCode := fn(c, &ambulance)
	if updatedAmbulance != nil {
		_, err := col.UpdateOne(ctx, bson.M{"_id": ambulanceId}, bson.M{"$set": updatedAmbulance})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to update ambulance",
			})
			return
		}
	}
	c.JSON(statusCode, result)
}

// CreateAmbulance handles POST /api/ambulances.
// It creates a new ambulance in MongoDB.
func (o *implAmbulanceAPI) CreateAmbulance(c *gin.Context) {
	var ambulance Ambulance
	if err := c.ShouldBindJSON(&ambulance); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}
	// Generate a new UUID if no ID is provided.
	if ambulance.Id == "" {
		ambulance.Id = uuid.NewString()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := o.collection.InsertOne(ctx, ambulance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Error inserting ambulance",
		})
		return
	}
	c.JSON(http.StatusCreated, ambulance)
}

// DeleteAmbulance handles DELETE /api/ambulances/:ambulanceId.
// It deletes the specified ambulance (and in a complete implementation, associated procedures) from MongoDB.
func (o *implAmbulanceAPI) DeleteAmbulance(c *gin.Context) {
	updateAmbulanceFunc(c, func(c *gin.Context, ambulance *Ambulance) (*Ambulance, interface{}, int) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := o.collection.DeleteOne(ctx, bson.M{"_id": ambulance.Id})
		if err != nil {
			return nil, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to delete ambulance",
			}, http.StatusInternalServerError
		}
		return nil, nil, http.StatusNoContent
	}, o.collection)
}

// GetAmbulanceById handles GET /api/ambulances/:ambulanceId.
// It retrieves and returns the details of a specific ambulance.
func (o *implAmbulanceAPI) GetAmbulanceById(c *gin.Context) {
	updateAmbulanceFunc(c, func(c *gin.Context, ambulance *Ambulance) (*Ambulance, interface{}, int) {
		return nil, ambulance, http.StatusOK
	}, o.collection)
}

// GetAmbulances handles GET /api/ambulances.
// It retrieves and returns a list of all ambulances from MongoDB.
func (o *implAmbulanceAPI) GetAmbulances(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := o.collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Error fetching ambulances",
		})
		return
	}
	defer cursor.Close(ctx)

	var ambulances []Ambulance
	if err := cursor.All(ctx, &ambulances); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Error decoding ambulances",
		})
		return
	}
	c.JSON(http.StatusOK, ambulances)
}

// UpdateAmbulance handles PUT /api/ambulances/:ambulanceId.
// It updates the details of an existing ambulance in MongoDB.
func (o *implAmbulanceAPI) UpdateAmbulance(c *gin.Context) {
	updateAmbulanceFunc(c, func(c *gin.Context, ambulance *Ambulance) (*Ambulance, interface{}, int) {
		var updated Ambulance
		if err := c.ShouldBindJSON(&updated); err != nil {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid request body",
				"error":   err.Error(),
			}, http.StatusBadRequest
		}

		// Update ambulance fields if provided in the request.
		// Adjust the fields based on your model definition.
		if updated.Name != "" {
			ambulance.Name = updated.Name
		}
		if updated.Location != "" {
			ambulance.Location = updated.Location
		}
		if updated.DriverName != "" {
			ambulance.DriverName = updated.DriverName
		}
		// Add more field updates as necessary.

		return ambulance, ambulance, http.StatusOK
	}, o.collection)
}

// GetAmbulanceSummary handles GET /api/ambulances/:ambulanceId/summary.
// It returns a summary of procedure costs for an ambulance.
// For demonstration purposes, this returns a dummy summary.
func (o *implAmbulanceAPI) GetAmbulanceSummary(c *gin.Context) {
	updateAmbulanceFunc(c, func(c *gin.Context, ambulance *Ambulance) (*Ambulance, interface{}, int) {
		summary := gin.H{
			"ambulanceId": ambulance.Id,
			"totalCost":   1500.50,
		}
		return nil, summary, http.StatusOK
	}, o.collection)
}
