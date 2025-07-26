package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
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

	// Issue a new access token if refresh token is valid
	newAccessToken, err := generateJWT(user, JWT_SECRET)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate new access token"})
		return
	}

	// Set new access token in cookie
	c.SetCookie("access_token", newAccessToken, COOKIE_OPTIONS.MaxAge, "/", "", COOKIE_OPTIONS.Secure, COOKIE_OPTIONS.HttpOnly)

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
