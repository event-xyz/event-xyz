package forms

import (
	"fmt"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/eventloop-testbed/backend/db"
	"github.com/gin-gonic/gin"
)

type FormWrapper struct {
	Form Form `json:"forms"`
}

// GetForm handler to fetch the form by docID
func GetForm(c *gin.Context) {
	docID := c.Param("docID")

	// Get the scope to query from the 'eventloop' scope
	scope := db.InitialiseBucket().Scope("eventloop")

	// Define the query to retrieve the form by docID
	query := "SELECT * FROM `forms` WHERE id = $1 LIMIT 1;"
	rows, err := scope.Query(query, &gocb.QueryOptions{
		Adhoc:                true,
		PositionalParameters: []interface{}{docID},
		Timeout:              15 * time.Second,
	})
	if err != nil {
		// Return an error response if the query fails
		fmt.Printf("Error executing query: %v\n", err)
		c.JSON(500, gin.H{"error": "Failed to execute query"})
		return
	}

	var form FormWrapper

	// Check if rows exist and scan the result into the Form struct
	if rows.Next() {
		if err := rows.Row(&form); err == nil {
			// If no error while scanning, return the form in the response
			c.JSON(200, gin.H{
				"id":     form.Form.ID,
				"title":  form.Form.Title,
				"fields": form.Form.Fields,
			})
			return
		}
	}

	// If no form found for docID, return a not found error
	c.JSON(404, gin.H{"error": "Form not found"})
}
