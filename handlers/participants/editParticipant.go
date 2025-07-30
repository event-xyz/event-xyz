package participants

import (
	"fmt"
	"net/http"

	"github.com/eventloop-testbed/backend/db"

	"github.com/gin-gonic/gin"
)

func EditParticipant(c *gin.Context) {
	var participant struct {
		ID          string `json:"id"`
		TeamID      string `json:"team_id"` // Reference to Team ID
		Role        string `json:"role"`
		Name        string `json:"name"`
		Email       string `json:"email"`
		Phone       string `json:"phone"`
		College     string `json:"college"`
		SRN         string `json:"srn"`
		Branch      string `json:"branch"`
		DayScholar  bool   `json:"day_scholar"`      // Bool value (true/false)
		Hostel      string `json:"hostel,omitempty"` // Nullable field
		Shortlisted bool   `json:"shortlisted"`      // Bool value (true/false)
		QRString    string `json:"qr_string"`
	}

	if err := c.ShouldBindJSON(&participant); err != nil {
		fmt.Printf("json bind err: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	var existingParticipant struct {
		ID          string `json:"id"`
		TeamID      string `json:"team_id"` // Reference to Team ID
		Role        string `json:"role"`
		Name        string `json:"name"`
		Email       string `json:"email"`
		Phone       string `json:"phone"`
		College     string `json:"college"`
		SRN         string `json:"srn"`
		Branch      string `json:"branch"`
		DayScholar  bool   `json:"day_scholar"`      // Bool value (true/false)
		Hostel      string `json:"hostel,omitempty"` // Nullable field
		Shortlisted bool   `json:"shortlisted"`      // Bool value (true/false)
		QRString    string `json:"qr_string"`
	}

	collection := db.InitialiseBucket().Scope("eventloop").Collection("participants")

	getRes, err := collection.Get(participant.ID, nil)
	if err != nil {
		fmt.Printf("Error retrieving existing participant: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := getRes.Content(&existingParticipant); err != nil {
		fmt.Printf("Error copying 'existingParticipant': %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update only the fields that were provided (non-empty fields)
	if participant.Name != "" {
		existingParticipant.Name = participant.Name
	}
	if participant.Email != "" {
		existingParticipant.Email = participant.Email
	}
	if participant.Role != "" {
		existingParticipant.Role = participant.Role
	}
	if participant.Phone != "" {
		existingParticipant.Phone = participant.Phone
	}
	if participant.College != "" {
		existingParticipant.College = participant.College
	}
	if participant.SRN != "" {
		existingParticipant.SRN = participant.SRN
	}
	if participant.Branch != "" {
		existingParticipant.Branch = participant.Branch
	}
	if participant.DayScholar {
		existingParticipant.DayScholar = participant.DayScholar
	}
	if participant.Hostel != "" {
		existingParticipant.Hostel = participant.Hostel
	}
	if participant.Shortlisted {
		existingParticipant.Shortlisted = participant.Shortlisted
	}
	if participant.QRString != "" {
		existingParticipant.QRString = participant.QRString
	}

	// Insert the user document into the Capella collection
	mutRes, err := collection.Upsert(participant.ID, existingParticipant, nil)

	if err != nil {
		fmt.Printf("internal server err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the success response with the document ID
	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"mutationResult": mutRes,
		"user":           existingParticipant,
	})
}
