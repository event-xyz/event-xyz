package handlers

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/api/idtoken"
)

func VerifyToken(idToken string) (map[string]interface{}, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID not set")
	}

	payload, err := idtoken.Validate(context.Background(), idToken, clientID)
	if err != nil {
		return nil, fmt.Errorf("unable to verify token: %v", err)
	}

	return map[string]interface{}{
		"name":  payload.Claims["name"],
		"email": payload.Claims["email"],
	}, nil
}
