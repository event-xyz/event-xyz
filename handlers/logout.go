package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func Logout(c *gin.Context) {
	cookieOptions := &http.Cookie{
		HttpOnly: true,
		Secure:   os.Getenv("PRODUCTION") == "true",
	}

	c.SetCookie("access_token", "", -1, "/", "", cookieOptions.Secure, cookieOptions.HttpOnly)
	c.SetCookie("refresh_token", "", -1, "/refresh", "", cookieOptions.Secure, cookieOptions.HttpOnly)

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully.",
	})
}
