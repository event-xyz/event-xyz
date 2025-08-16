package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/eventloop-testbed/backend/db"
	"github.com/eventloop-testbed/backend/models"

	"github.com/couchbase/gocb/v2"
)

// AuthUserResult holds the result structure
type AuthUserResult struct {
	Success   bool   `json:"success"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	QR_string string `json:"qr_string"`
}

// GetAuthUser implements the logic to fetch or create a user
func GetAuthUser(name, email string) (*AuthUserResult, error) {
	bucket := db.InitialiseBucket()
	scope := bucket.Scope("eventloop")
	partCol := scope.Collection("participants")

	// 1. Try to find user in dbAuthorisedUsers
	query := "SELECT * FROM `dbAuthorisedUsers` WHERE email = $1 LIMIT 1;"
	rows, err := scope.Query(query, &gocb.QueryOptions{
		Adhoc:                true,
		PositionalParameters: []interface{}{email},
		Timeout:              15 * time.Second,
	})

	if err != nil {
		fmt.Printf("%v\n", email)
		fmt.Printf("err occured: %v", err)
		return nil, err
	}
	var dbAuthUser struct {
		DbAuthorisedUsers struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
			Role  string `json:"role"`
		} `json:"dbAuthorisedUsers"`
	}

	if rows.Next() {
		if err := rows.Row(&dbAuthUser); err == nil {
			fmt.Printf("no errors getting records")
			fmt.Print(dbAuthUser)
			return &AuthUserResult{Success: true, ID: dbAuthUser.DbAuthorisedUsers.ID, Name: dbAuthUser.DbAuthorisedUsers.Name, Email: dbAuthUser.DbAuthorisedUsers.Email, Role: dbAuthUser.DbAuthorisedUsers.Role}, nil
		}
	}
	// 2. Try to find user in participants
	query = "SELECT * FROM `participants` WHERE email = $1 LIMIT 1"
	rows, err = scope.Query(query, &gocb.QueryOptions{
		Adhoc:                true,
		PositionalParameters: []interface{}{email},
		Timeout:              15 * time.Second,
	})

	if err != nil {
		return nil, err
	}

	var user struct {
		Participant struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Email     string `json:"email"`
			Role      string `json:"role"`
			QR_string string `json:"qr_string"`
		} `json:"participants"`
	}

	if rows.Next() {
		if err := rows.Row(&user); err == nil {
			return &AuthUserResult{Success: true, ID: user.Participant.ID, Name: user.Participant.Name, Email: user.Participant.Email, Role: user.Participant.Role, QR_string: user.Participant.QR_string}, nil
		}
	}
	// 3. Insert new participant
	newParticipant := models.Participant{
		ID:          GenerateDocID(email),
		Name:        name,
		Email:       email,
		Role:        "participant",
		Phone:       "NA",
		College:     "NA",
		SRN:         "NA",
		Branch:      "NA",
		DayScholar:  false,
		Hostel:      "NA",
		Shortlisted: false,
		QRString:    "NA",
	}

	encodedString, qrCodeBase64, err := GenerateQRCode(newParticipant.ID, os.Getenv("QR_SECRET_KEY"))

	newParticipant.QRString = encodedString

	if err != nil {
		return nil, err
	}
	_, upsetErr := partCol.Upsert(newParticipant.ID, newParticipant, &gocb.UpsertOptions{Timeout: 15 * time.Second})

	if upsetErr != nil {
		return nil, err
	}
	return &AuthUserResult{Success: true, ID: newParticipant.ID, Name: newParticipant.Name, Email: newParticipant.Email, Role: "participant", QR_string: qrCodeBase64}, nil
}
