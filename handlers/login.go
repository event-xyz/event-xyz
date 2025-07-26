package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/eventloop-testbed/backend/db/utils"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var JWT_SECRET = os.Getenv("JWT_SECRET")
var REFRESH_JWT_SECRET = os.Getenv("REFRESH_JWT_SECRET")
var COOKIE_OPTIONS = &http.Cookie{
	HttpOnly: true,
	Secure:   os.Getenv("PRODUCTION") == "true",
	SameSite: http.SameSiteLaxMode,
	MaxAge:   24 * 60 * 60,
}

func Login(c *gin.Context) {
	var request struct {
		Credentials string `json:"credentials"`
	}

	// Parse JSON body
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing credentials"})
		return
	}

	// Verify the Google ID token (you'll implement this using Google OAuth in verifyToken function)
	user, err := VerifyToken(request.Credentials)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token or server error"})
		return
	}

	// Extract user details
	email := user["email"].(string)
	name := user["name"].(string)

	// For now, we're mocking the result of database lookup
	result, err := utils.GetAuthUser(name, email)

	// If no user found
	if result == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "403 forbidden, records missing"})
		return
	}

	// Create JWT access and refresh tokens
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"name":  name,
		"role":  result.Role,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})
	accessTokenString, _ := accessToken.SignedString([]byte(JWT_SECRET))

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"name":  name,
		"role":  result.Role,
		"exp":   time.Now().Add(72 * time.Hour).Unix(),
	})
	refreshTokenString, _ := refreshToken.SignedString([]byte(REFRESH_JWT_SECRET))

	// Set cookies
	c.SetCookie("access_token", accessTokenString, 24*60*60, "/", "", COOKIE_OPTIONS.Secure, true)
	c.SetCookie("refresh_token", refreshTokenString, 72*60*60, "/refresh", "", COOKIE_OPTIONS.Secure, true)

	// Respond with user info
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    result.ID,
			"name":  name,
			"email": email,
			"role":  result.Role,
		},
	})
}
