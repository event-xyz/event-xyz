package events

import (
	"log"
	"net/http"
	"os"

	"github.com/eventloop-testbed/backend/db"
	"github.com/eventloop-testbed/backend/db/utils"
	"github.com/eventloop-testbed/backend/handlers/forms"
	"github.com/eventloop-testbed/backend/models" // adjust the import path as needed
	"github.com/gin-gonic/gin"
)

// CreateEvent handles the creation of a new event
func CreateEvent(c *gin.Context) {
	var newEvent models.Events

	// Bind the incoming JSON payload to the Events struct
	if err := c.ShouldBindJSON(&newEvent); err != nil {
		log.Printf("error invalid req data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	// Here you can add validation checks for event fields if needed
	log.Printf("event date: %v", newEvent.Event_Date)

	if newEvent.Name == "" || newEvent.Event_Date == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Name and event date are required",
		})
		return
	}

	newEvent.ID = utils.GenerateDocID(newEvent.Name)

	// create an empty form to link the event to
	emptyField := forms.Field{
		Label:    "Comments",
		Type:     "long_text",
		Options:  []string{"value"},
		Required: false,
	}

	emptyForm := forms.Form{
		Title:  newEvent.Name,   // Use the event name as the form title
		Fields: []forms.Field{}, // Empty fields
	}

	emptyForm.Fields = append(emptyForm.Fields, emptyField)

	// Save the empty form to the database
	formDocID := utils.GenerateDocID(newEvent.Name)
	emptyForm.ID = formDocID

	formCollection := db.InitialiseBucket().Scope("eventloop").Collection("forms")
	_, err := formCollection.Upsert(formDocID, emptyForm, nil)
	if err != nil {
		log.Printf("Failed to save form: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create associated form"})
		return
	}

	newEvent.Form_Link = os.Getenv("FRONTEND_ENDPOINT") + "forms/" + formDocID + "/submit"

	collection := db.InitialiseBucket().Scope("eventloop").Collection("events")
	_, err = collection.Upsert(newEvent.ID, newEvent, nil)
	if err != nil {
		log.Printf("Failed to save form: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save form"})
		return
	}

	// Respond with success and the IDs of the created event and form
	c.JSON(http.StatusCreated, gin.H{
		"message":  "Event created!",
		"event_id": newEvent.ID,
		"form_id":  formDocID,
	})
}
