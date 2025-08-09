package qrscan

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/dgrijalva/jwt-go"
	"github.com/eventloop-testbed/backend/db"
	"github.com/gin-gonic/gin"
)

func QRScan(c *gin.Context) {
	log.Printf("entered func")

	// Retrieve user from middleware (should be set in context)
	user, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "not authenticated"})
		return
	}

	claims, ok := user.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid user claims"})
		return
	}

	// Only admins allowed
	if claims["role"] != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "teri fielding set hai",
		})
		return
	}

	// Parse qr_string from JSON body
	var req struct {
		QRString string `json:"qr_string"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "missing qr_string"})
		return
	}

	scope := db.InitialiseBucket().Scope("eventloop")

	var participant struct {
		Participant struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Email     string `json:"email"`
			Role      string `json:"role"`
			QR_string string `json:"qr_string"`
		} `json:"participants"`
	}

	log.Printf("req qr string: %v", req.QRString)

	query := "SELECT * from `participants` WHERE qr_string = $1 LIMIT 1;"
	rows, err := scope.Query(query, &gocb.QueryOptions{
		Adhoc:                true,
		PositionalParameters: []interface{}{req.QRString},
		Timeout:              10 * time.Second,
	})

	if rows.Next() {
		if err := rows.Row(&participant); err == nil {
			// Compute expected QR string
			rawString := fmt.Sprintf("%s:%s", participant.Participant.Email, os.Getenv("QR_SECRET_KEY"))
			hash := sha256.New()
			hash.Write([]byte(rawString))
			expectedQRString := base64.URLEncoding.EncodeToString(hash.Sum(nil))

			log.Printf("AUTH: claims[email]=%v", participant.Participant.Email)
			log.Printf("AUTH: expected=%s", expectedQRString)
			log.Printf("AUTH: received=%s", req.QRString)

			if expectedQRString == req.QRString {
				c.JSON(http.StatusOK, gin.H{
					"message": "Authenticated QR",
					"user":    participant.Participant, // send only claims, not full user struct
				})
				return
			}
		}
	}

	if err != nil {
		fmt.Printf("err occured: %v", err)
	}

	c.JSON(http.StatusUnauthorized, gin.H{
		"message": "Invalid QR",
	})
}
