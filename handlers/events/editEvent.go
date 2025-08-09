package events

import (
	"log"
	"net/http"
	"net/url"

	"github.com/eventloop-testbed/backend/db"
	"github.com/eventloop-testbed/backend/models"
	"github.com/gin-gonic/gin"
)

// UpdateForm handles updating an existing form in the database
func UpdateForm(c *gin.Context) {
	var event models.Events

	// Bind the incoming JSON body to the Form struct
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the docID from the URL parameter
	docID := c.Param("docID")

	decodedEventID, error := url.QueryUnescape(event.ID)
	if error != nil {
		log.Printf("Failed to decode docID: %v", error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode event ID"})
		return
	}

	// Check if the provided event ID matches the docID in the URL
	if decodedEventID != docID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event ID does not match URL docID"})
		return
	}

	// Get the Couchbase collection for events
	collection := db.InitialiseBucket().Scope("eventloop").Collection("events")

	// Update the event in Couchbase using Upsert (this will insert the form if it doesn't exist)
	_, err := collection.Replace(decodedEventID, event, nil)
	if err != nil {
		log.Printf("Failed to update event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "event updated successfully", "event_id": decodedEventID})
}
