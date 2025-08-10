package handlers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/dgrijalva/jwt-go"
	"github.com/eventloop-testbed/backend/db"
	"github.com/eventloop-testbed/backend/db/utils"
	"github.com/gin-gonic/gin"
)

func Refresh(c *gin.Context) {
	accessToken, err := c.Cookie("access_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing access token"})
		return
	}

	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}

	// Verify access token
	user, err := verifyJWT(accessToken, JWT_SECRET)
	if err == nil {
		qrString, err := getQRCodeString(user["email"].(string))

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch qr string"})
			return
		}

		user["qr_string"] = qrString
		c.JSON(http.StatusOK, gin.H{"loggedIn": true, "user": user})
		return
	}

	// If access token is expired or invalid, try to verify refresh token
	user, err = verifyJWT(refreshToken, REFRESH_JWT_SECRET)
	if err != nil {
		c.SetCookie("access_token", "", -1, "/", "", COOKIE_OPTIONS.Secure, COOKIE_OPTIONS.HttpOnly)
		c.SetCookie("refresh_token", "", -1, "/", "", COOKIE_OPTIONS.Secure, COOKIE_OPTIONS.HttpOnly)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expired"})
		return
	}

	qrString, err := getQRCodeString(user["email"].(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch qr string"})
		return
	}

	// Issue a new access token if refresh token is valid
	newAccessToken, err := generateJWT(user, JWT_SECRET)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate new access token"})
		return
	}

	// Set new access token in cookie
	c.SetCookie("access_token", newAccessToken, COOKIE_OPTIONS.MaxAge, "/", "", COOKIE_OPTIONS.Secure, COOKIE_OPTIONS.HttpOnly)

	user["qr_string"] = qrString

	c.JSON(http.StatusOK, gin.H{"loggedIn": true, "user": user})
}

func verifyJWT(tokenStr, secret string) (map[string]interface{}, error) {
	// Parse and validate JWT
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	// If token is valid, extract the user info
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

func generateJWT(user map[string]interface{}, secret string) (string, error) {
	// Generate a new JWT with the user data
	claims := jwt.MapClaims{}
	for key, value := range user {
		claims[key] = value
	}
	claims["exp"] = time.Now().Add(24 * time.Hour).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func getQRCodeString(email string) (string, error) {
	// Perform the database query to retrieve the participant and their QR string
	query := "SELECT * FROM `participants` WHERE email = $1 LIMIT 1"
	rows, err := db.InitialiseBucket().Scope("eventloop").Query(query, &gocb.QueryOptions{
		Adhoc:                true,
		PositionalParameters: []interface{}{email},
		Timeout:              15 * time.Second,
	})

	if err != nil {
		return "", err
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
		if err := rows.Row(&user); err != nil {
			return "", err
		}

		// If QR string doesn't exist in the DB, generate and store it
		if user.Participant.QR_string == "" {
			encodedString, qrCodeBase64, err := utils.GenerateQRCode(user.Participant.ID, os.Getenv("QR_SECRET_KEY"))
			if err != nil {
				return "", err
			}

			// Store the base64-encoded QR string in the database
			partCol := db.InitialiseBucket().Collection("participants")
			_, err = partCol.Replace(user.Participant.ID, map[string]interface{}{
				"qr_string": encodedString, // Store the base64-encoded QR code
			}, &gocb.ReplaceOptions{
				Timeout: 15 * time.Second,
			})

			if err != nil {
				return "", err
			}

			// Return the qrCodeBase64 for frontend rendering
			return qrCodeBase64, nil
		}

		// If QR string exists, return the qrCodeBase64 for frontend rendering
		return user.Participant.QR_string, nil
	}

	return "", fmt.Errorf("user not found")
}
