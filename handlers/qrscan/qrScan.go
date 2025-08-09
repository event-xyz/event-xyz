package qrscan

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func QRScan(c *gin.Context) {
	log.Printf("entered func")
	user, ok := c.Get("user")
	if !ok {
		fmt.Print(ok)
	}

	if user.(jwt.MapClaims)["role"] == "admin" {
		givenQRString, ok := c.Get("qr_string")

		if !ok {
			fmt.Print(ok)
		}

		rawString := fmt.Sprintf("%s:%s", user.(jwt.MapClaims)["email"], os.Getenv("QR_SECRET_KEY"))

		// Hash the concatenated string for better security
		hash := sha256.New()
		hash.Write([]byte(rawString))
		encodedString := base64.URLEncoding.EncodeToString(hash.Sum(nil))

		log.Printf("auth success")

		if encodedString == givenQRString {
			c.JSON(http.StatusOK, gin.H{
				"message": "Authenticated QR",
				"user":    user,
			})
			return
		}
		return
	} else {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "teri fielding set hai",
		})
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"message": "very bad thing happened",
	})
}
