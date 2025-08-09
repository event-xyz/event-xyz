package forms

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/eventloop-testbed/backend/db"
	"github.com/eventloop-testbed/backend/db/utils"
	"github.com/gin-gonic/gin"
)

type FormResponse struct {
	ID       string            `json:"id"`
	FormID   string            `json:"docID"`
	Response map[string]string `json:"response"`
}

func SubmitResponse(c *gin.Context) {
	user, ok := c.Get("user")
	if !ok {
		fmt.Print(ok)
	}

	docID := c.Param("docID") // Fetch the docID from the URL parameter

	var formResponse FormResponse
	if err := c.ShouldBindJSON(&formResponse); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	user = user.(jwt.MapClaims)["email"].(string)

	// Create a new document to store responses in Couchbase
	responseDocID := utils.GenerateDocID(user.(string)) // Create a unique document ID

	// Prepare the document
	document := map[string]interface{}{
		"ID":       responseDocID,
		"docID":    docID,
		"response": formResponse.Response,
	}

	collection := db.InitialiseBucket().Scope("eventloop").Collection("form_response")

	// Insert the document into the form_responses collection
	_, err := collection.Upsert(responseDocID, document, nil)
	if err != nil {
		log.Printf("Failed to insert document into Couchbase: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit your response"})
		return
	}

	// Return a success response
	c.JSON(http.StatusOK, gin.H{"message": "Response submitted successfully!"})
}
