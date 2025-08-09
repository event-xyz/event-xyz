package events

import (
	"fmt"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/dgrijalva/jwt-go"
	"github.com/eventloop-testbed/backend/db"
	"github.com/gin-gonic/gin"
)

func GetAllEvents(c *gin.Context) {
	user, ok := c.Get("user")
	if !ok {
		fmt.Print(ok)
	}

	scope := db.InitialiseBucket().Scope("eventloop")

	if user.(jwt.MapClaims)["role"].(string) == "participant" {
		email := user.(jwt.MapClaims)["email"].(string)
		// query the db for event details for which the participant has submitted their form
		// if atleast 1 character from both strings are removed, then it allows for partial string matching. else, it will try to match every character down to the milliseconds which isn't possible.
		query := `SELECT events.id, events.name FROM events JOIN form_response ON SUBSTR(events.id, 0, LENGTH(events.id) - 6) = SUBSTR(form_response.docID, 0, LENGTH(form_response.docID) - 6) WHERE CONTAINS(form_response.ID, $1);`

		rows, err := scope.Query(query, &gocb.QueryOptions{
			Adhoc:                true,
			PositionalParameters: []interface{}{email},
			Timeout:              5 * time.Second,
		})
		if err != nil {
			// Return an error response if the query fails
			fmt.Printf("Error executing query: %v\n", err)
			c.JSON(500, gin.H{"error": "Failed to execute query"})
			return
		}

		var events []map[string]interface{}
		// Check if rows exist and scan the result into the Form struct
		for rows.Next() {
			var event map[string]interface{}
			// Scan the row into the form struct
			if err := rows.Row(&event); err != nil {
				fmt.Printf("Error scanning row: %v\n", err)
				continue
			}

			// Append the scanned form to the forms slice
			events = append(events, event)
		}

		c.JSON(200, gin.H{
			"events": events,
		})
		return
	}

	// Get all events
	query := "SELECT id, name FROM events;"

	rows, err := scope.Query(query, &gocb.QueryOptions{
		Adhoc:                true,
		PositionalParameters: []interface{}{},
		Timeout:              5 * time.Second,
	})
	if err != nil {
		// Return an error response if the query fails
		fmt.Printf("Error executing query: %v\n", err)
		c.JSON(500, gin.H{"error": "Failed to execute query"})
		return
	}

	var events []map[string]interface{}
	// Check if rows exist and scan the result into the Form struct
	for rows.Next() {
		var event map[string]interface{}
		// Scan the row into the event struct
		if err := rows.Row(&event); err != nil {
			fmt.Printf("Error scanning row: %v\n", err)
			continue
		}

		// Append the scanned event to the events slice
		events = append(events, event)
	}

	c.JSON(200, gin.H{
		"events": events,
	})
}
