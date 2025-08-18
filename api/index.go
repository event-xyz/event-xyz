package handler

import (
	"fmt"
	"log"
	"os"

	"github.com/eventloop-testbed/backend/db"
	"github.com/eventloop-testbed/backend/handlers"
	"github.com/eventloop-testbed/backend/handlers/dbAuthorisedUsers"
	"github.com/eventloop-testbed/backend/handlers/events"
	"github.com/eventloop-testbed/backend/handlers/forms"
	"github.com/eventloop-testbed/backend/handlers/participants"
	"github.com/eventloop-testbed/backend/handlers/qrscan"
	"github.com/eventloop-testbed/backend/handlers/teams"
	"github.com/eventloop-testbed/backend/middleware"
	"github.com/joho/godotenv"

	"net/http"

	"github.com/gin-gonic/gin"
)

const red = "\033[31m"

var app *gin.Engine

func apiRouter(r *gin.RouterGroup) {
	r.POST("/dbAuthorisedUsers/add", middleware.ValidateUserDetails(), dbAuthorisedUsers.AddUser)

	r.POST("/teams/create", teams.AddTeam)

	r.GET("/refresh", handlers.Refresh)
	r.POST("/logout", handlers.Logout)

	r.POST("/participants/add", participants.AddParticipant)
	r.POST("/participants/edit", participants.EditParticipant)

	r.GET("/events/", events.GetAllEvents)
	r.POST("/events/create", events.CreateEvent)
	r.OPTIONS("/events/create", events.CreateEvent)

	r.POST("/forms/create", forms.SaveForm)
	r.GET("/forms/", forms.GetAllForms)
	r.GET("/forms/:docID", forms.GetForm)
	r.PUT("/forms/:docID/edit", forms.UpdateForm)
	r.OPTIONS("/forms/:docID/edit", forms.UpdateForm)

	r.GET("/forms/:docID/responses", forms.Responses)
	r.POST("/forms/:docID/submit_response", forms.SubmitResponse)

	r.POST("/verifyQR", qrscan.QRScan)
	r.OPTIONS("/verifyQR", qrscan.QRScan)
}

func Main() {
	app = gin.New()
	r := app.Group("/")
	r.Use(middleware.CorsMiddleware())

	// Prevents use of `SetUser` middleware on these routes
	r.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ping OK, server running...")
	})

	r.POST("/login", handlers.Login)
	r.OPTIONS("/login", handlers.Login)

	r.Use(middleware.SetUser())

	bucket := db.InitialiseBucket().Scope("eventloop")
	if bucket == nil {
		log.Fatalf(red + "Failed to initialise bucket")
	} else {
		fmt.Printf("Bucket '%s' initialised\n", bucket.Name())
	}

	apiRouter(r)
}

// This function will check whether all env variables are defined.
// If any one of them is missing, it will throw an appropriate error and exit
func checkEnvVars() {
	requiredEnvVars := []string{
		"PRODUCTION",
		"DB_CONNECTION_STRING",
		"DB_USERNAME",
		"DB_PASSWORD",
		"DB_BUCKET_NAME",
		"JWT_SECRET",
		"REFRESH_JWT_SECRET",
		"QR_SECRET_KEY",
		"GOOGLE_CLIENT_ID",
		"FRONTEND_ENDPOINT",
	}

	_, err := os.Stat(".env")
	if err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Printf("Error loading .env file")
		}
	}

	for _, envVar := range requiredEnvVars {
		value, exists := os.LookupEnv(envVar)
		if !exists || value == "" {
			log.Fatalf(red+"ERROR: Environment variable '%s' is missing.", envVar)
		}
	}

	fmt.Println("Starting server...")
}

func Handler(w http.ResponseWriter, r *http.Request) {
	checkEnvVars()

	if app == nil {
		Main()
	}
	app.ServeHTTP(w, r)
}
