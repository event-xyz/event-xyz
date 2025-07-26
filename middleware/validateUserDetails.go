package middleware

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

// ValidateUserDetails validates the user details in the request body
func ValidateUserDetails() gin.HandlerFunc {
	return func(c *gin.Context) {
		var userInput struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Role  string `json:"role"`
		}

		// Bind the incoming JSON to the userInput struct
		if err := c.ShouldBindJSON(&userInput); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			c.Abort()
			return
		}

		// Validate name
		if len(userInput.Name) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
			c.Abort()
			return
		}

		// Validate email (basic format check)
		if !isValidEmail(userInput.Email) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
			c.Abort()
			return
		}

		roles := [3]string{"admin", "organiser", "volunteer"}

		if !includes(roles, userInput.Role) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
			c.Abort()
			return
		}

		// Attach validated data to the context
		c.Set("validatedName", userInput.Name)
		c.Set("validatedEmail", userInput.Email)
		c.Set("validatedRole", userInput.Role)

		c.Next()
	}
}

func includes(slice [3]string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// isValidEmail checks if the email format is valid using regex
func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}
