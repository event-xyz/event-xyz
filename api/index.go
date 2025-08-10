package handler

import (
	"fmt"

	"github.com/eventloop-testbed/backend/db"
	"github.com/eventloop-testbed/backend/handlers"
	"github.com/eventloop-testbed/backend/handlers/dbAuthorisedUsers"
	"github.com/eventloop-testbed/backend/handlers/events"
	"github.com/eventloop-testbed/backend/handlers/forms"
	"github.com/eventloop-testbed/backend/handlers/participants"
	"github.com/eventloop-testbed/backend/handlers/qrscan"
	"github.com/eventloop-testbed/backend/handlers/teams"
	"github.com/eventloop-testbed/backend/middleware"

	"net/http"

	"github.com/gin-gonic/gin"
)

var app *gin.Engine

func apiRouter(r *gin.RouterGroup) {
	r.GET("/health", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ping")
	})

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
	r.Use(handlers.CorsMiddleware())

	r.POST("/login", handlers.Login)
	r.OPTIONS("/login", handlers.Login)

	r.Use(middleware.SetUser())

	bucket := db.InitialiseBucket().Scope("eventloop")
	if bucket == nil {
		panic("Failed to initialise bucket")
	} else {
		fmt.Printf("Bucket '%s' initialised\n", bucket.Name())
	}

	apiRouter(r)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if app == nil {
		Main()
	}
	app.ServeHTTP(w, r)
}
