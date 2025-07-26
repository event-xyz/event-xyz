package teams

import (
	"net/http"
	"time"

	"github.com/eventloop-testbed/backend/db"
	"github.com/eventloop-testbed/backend/db/utils"
	"github.com/eventloop-testbed/backend/models"

	"github.com/gin-gonic/gin"
)

func AddTeam(c *gin.Context) {
	var userInput struct {
		ID              string    `json:"id"`
		EventID         int       `json:"event_id"`
		Name            string    `json:"name"`
		LeadName        string    `json:"lead_name"`
		LeadEmail       string    `json:"lead_email"`
		LeadPhone       string    `json:"lead_phone"`
		NumParticipants int       `json:"num_participants"`
		CreatedAt       time.Time `json:"created_at"`
		UpdatedAt       time.Time `json:"updated_at"`
	}

	// Bind the JSON data from the request body to the userInput struct
	if err := c.ShouldBindJSON(&userInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		c.Abort()
		return
	}

	team := models.Team{
		ID:              utils.GenerateDocID(userInput.Name),
		EventID:         userInput.EventID,
		Name:            userInput.Name,
		LeadName:        userInput.LeadName,
		LeadEmail:       userInput.LeadEmail,
		LeadPhone:       userInput.LeadPhone,
		NumParticipants: userInput.NumParticipants,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	collection := db.InitialiseBucket().Scope("eventloop").Collection("teams")

	mutRes, err := db.CreateDocument(collection, collection.Name(), team.ID, team)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the success response with the document ID
	c.JSON(http.StatusCreated, gin.H{
		"success":        true,
		"mutationResult": mutRes,
		"team":           team,
	})
}
