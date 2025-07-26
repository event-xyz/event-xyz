package main

import (
	"fmt"

	"github.com/eventloop-testbed/backend/db"
	"github.com/eventloop-testbed/backend/handlers"
	"github.com/eventloop-testbed/backend/handlers/dbAuthorisedUsers"
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

	r.POST("/login", handlers.Login)

	r.GET("/refresh", handlers.Refresh)

	r.POST("/logout", handlers.Logout)
}

func Main() {
	app = gin.New()
	r := app.Group("/")
	r.Use(handlers.CorsMiddleware())

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
