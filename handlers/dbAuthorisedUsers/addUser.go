package dbAuthorisedUsers

import (
	"net/http"

	"github.com/eventloop-testbed/backend/db"
	"github.com/eventloop-testbed/backend/db/utils"
	"github.com/eventloop-testbed/backend/models"

	"github.com/gin-gonic/gin"
)

// AddUser handles adding a new user
func AddUser(c *gin.Context) {
	// Validate and extract validated data from the context [ref: middleware (validateUserDetails)]
	name := c.MustGet("validatedName").(string)
	email := c.MustGet("validatedEmail").(string)
	role := c.MustGet("validatedRole").(string)

	// Create the user object using the model
	dbAuthorisedUser := models.DBAuthorisedUsers{
		ID:    utils.GenerateDocID(role),
		Name:  name,
		Email: email,
		Role:  role,
	}

	// Insert the user document into the Capella collection
	collection := db.InitialiseBucket().Scope("eventloop").Collection("dbAuthorisedUsers")

	mutRes, err := db.CreateDocument(collection, collection.Name(), dbAuthorisedUser.ID, dbAuthorisedUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the success response with the document ID
	c.JSON(http.StatusCreated, gin.H{
		"success":        true,
		"mutationResult": mutRes,
		"user":           dbAuthorisedUser,
	})
}
