package forms

import (
	"fmt"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/eventloop-testbed/backend/db"
	"github.com/gin-gonic/gin"
)

type FormResponseWrapper struct {
	Form FormResponse `json:"form_response"`
}

func Responses(c *gin.Context) {
	docID := c.Param("docID")

	// Get the scope to query from the 'eventloop' scope
	scope := db.InitialiseBucket().Scope("eventloop")

	// Define the query to retrieve the form by docID
	query := "SELECT ID, response FROM `form_response` WHERE docID = $1;"
	rows, err := scope.Query(query, &gocb.QueryOptions{
		Adhoc:                true,
		PositionalParameters: []interface{}{docID},
		Timeout:              5 * time.Second,
	})
	if err != nil {
		// Return an error response if the query fails
		fmt.Printf("Error executing query: %v\n", err)
		c.JSON(500, gin.H{"error": "Failed to execute query"})
		return
	}

	var responses []FormResponse

	// Check if rows exist and scan the result into the Form struct
	for rows.Next() {
		var formResponse FormResponse
		if err := rows.Row(&formResponse); err == nil {
			// Append the response
			responses = append(responses, formResponse)
		}
	}

	// If no responses found for docID, return a not found error
	if len(responses) == 0 {
		c.JSON(404, gin.H{"error": "No responses found"})
		return
	}

	// Return the responses
	c.JSON(200, gin.H{
		"formID":    docID,
		"responses": responses,
	})
}
