package forms

import (
	"fmt"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/dgrijalva/jwt-go"
	"github.com/eventloop-testbed/backend/db"
	"github.com/gin-gonic/gin"
)

func GetAllForms(c *gin.Context) {
	user, ok := c.Get("user")
	if !ok {
		fmt.Print(ok)
	}

	scope := db.InitialiseBucket().Scope("eventloop")

	if user.(jwt.MapClaims)["role"].(string) == "participant" {
		email := user.(jwt.MapClaims)["email"].(string)
		query := "SELECT ID, docID FROM `form_response` WHERE CONTAINS(ID, $1);"

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

		var forms []map[string]interface{}
		// Check if rows exist and scan the result into the Form struct
		for rows.Next() {
			var form map[string]interface{}
			// Scan the row into the form struct
			if err := rows.Row(&form); err != nil {
				fmt.Printf("Error scanning row: %v\n", err)
				continue
			}

			// Append the scanned form to the forms slice
			forms = append(forms, form)
		}

		c.JSON(200, gin.H{
			"forms": forms,
		})
		return
	}

	// Get all forms
	query := "SELECT id, title FROM forms;"

	rows, err := scope.Query(query, &gocb.QueryOptions{
		Adhoc:                true,
		PositionalParameters: []interface{}{},
		Timeout:              10 * time.Second,
	})
	if err != nil {
		// Return an error response if the query fails
		fmt.Printf("Error executing query: %v\n", err)
		c.JSON(500, gin.H{"error": "Failed to execute query"})
		return
	}

	var forms []map[string]interface{}
	// Check if rows exist and scan the result into the Form struct
	for rows.Next() {
		var form map[string]interface{}
		// Scan the row into the form struct
		if err := rows.Row(&form); err != nil {
			fmt.Printf("Error scanning row: %v\n", err)
			continue
		}

		// Append the scanned form to the forms slice
		forms = append(forms, form)
	}

	c.JSON(200, gin.H{
		"forms": forms,
	})
}
