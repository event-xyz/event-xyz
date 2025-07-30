package participants

import (
	"net/http"

	"github.com/eventloop-testbed/backend/db"
	"github.com/eventloop-testbed/backend/db/utils"
	"github.com/eventloop-testbed/backend/models"

	"github.com/gin-gonic/gin"
)

func AddParticipant(c *gin.Context) {
	// Validate and extract validated data from the context [ref: middleware (validateUserDetails)]
	name := c.MustGet("validatedName").(string)
	email := c.MustGet("validatedEmail").(string)
	role := "participant"

	// Create the user object using the model
	participant := models.Participant{
		ID:          utils.GenerateDocID(role),
		Name:        name,
		Email:       email,
		Role:        role,
		Phone:       "NA",
		College:     "NA",
		SRN:         "NA",
		Branch:      "NA",
		DayScholar:  false,
		Hostel:      "NA",
		Shortlisted: false,
		QRString:    "NA",
	}

	// Insert the user document into the Capella collection
	collection := db.InitialiseBucket().Scope("eventloop").Collection("participants")

	mutRes, err := db.CreateDocument(collection, collection.Name(), participant.ID, participant)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the success response with the document ID
	c.JSON(http.StatusCreated, gin.H{
		"success":        true,
		"mutationResult": mutRes,
		"user":           participant,
	})
}
