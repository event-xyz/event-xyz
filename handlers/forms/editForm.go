package forms

import (
	"log"
	"net/http"
	"net/url"

	"github.com/eventloop-testbed/backend/db"
	"github.com/gin-gonic/gin"
)

// UpdateForm handles updating an existing form in the database
func UpdateForm(c *gin.Context) {
	var form Form

	// Bind the incoming JSON body to the Form struct
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the docID from the URL parameter
	docID := c.Param("docID")

	decodedFormID, error := url.QueryUnescape(form.ID)
	if error != nil {
		log.Printf("Failed to decode docID: %v", error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode form ID"})
		return
	}

	// Check if the provided form ID matches the docID in the URL
	if decodedFormID != docID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Form ID does not match URL docID"})
		return
	}

	// Get the Couchbase collection for forms
	collection := db.InitialiseBucket().Scope("eventloop").Collection("forms")

	// Update the form in Couchbase using Upsert (this will insert the form if it doesn't exist)
	_, err := collection.Replace(decodedFormID, form, nil)
	if err != nil {
		log.Printf("Failed to update form: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update form"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Form updated successfully", "form_id": decodedFormID})
}
