// deadline: in 2-weeks Infinite
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/homebrew-ec-foss/eventloop/handlers"
)

func main() {
	isDevBuild := os.Getenv("ELOOP_DEV")
	isHttpsServer := os.Getenv("ELOOP_HTTP")

	app := handlers.InitializeAppWithConfig("./data/config2.json")

	r := gin.Default()
	r.Use(handlers.CorsMiddleware())

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "pong")
	})

	// Generic functions for admin control

	r.GET("/search", func(ctx *gin.Context) {})

	// fetch participants details based on JWT ID
	// fetched based on a qr code
	r.POST("/qrsearch", app.HandleQRFetch)

	r.GET("/participant", app.HandleParticipantFetch)
	// r.POST("/participant", handlers.HandleParticipantUpdate)

	// Additional team addition besides CSV
	// future prospect
	r.POST("/createteam", func(ctx *gin.Context) {})
	r.POST("/login", app.HandleLogin)

	////////////////////////////////////////////////

	// Endpoints accessed during events
	// eg: Crossing checkpoints, etc.

	if isDevBuild == "1" {
		r.POST("/test/creation", app.HandleCreateTest)
		r.PUT("/checkin", app.HandleCheckin)
		r.PUT("/checkout", app.HandleCheckout)
		r.PUT("/checkpoint", app.HandleCheckpoint)
		r.POST("/add/volunteer", app.HandleVolunteers)
		r.POST("/add/volunteerMan", app.HandleVolunteerManual)
	}

	// TODO: Handle checking by scanner
	// Router Groups for
	// 	- volunteers
	//  - organiser
	// 	- admin

	// NOTE:
	volunteers := r.Group("/volunteer", app.AuthenticationMiddleware("volunteer"))
	{
		// NOTE: endpoints active during events
		volunteers.PUT("/checkin", app.HandleCheckin)
		volunteers.PUT("/checkout", app.HandleCheckout)
		volunteers.PUT("/checkpoint", app.HandleCheckpoint)
	}

	// NOTE:
	organiser := r.Group("/organiser", app.AuthenticationMiddleware("organiser"))
	{

		// NOTE: endpoints active during events
		organiser.PUT("/checkin", app.HandleCheckin)
		organiser.PUT("/checkout", app.HandleCheckout)
		organiser.PUT("/checkpoint", app.HandleCheckpoint)
	}

	// NOTE:
	admin := r.Group("/admin", app.AuthenticationMiddleware("admin"))
	{

		// NOTE: endpoints active during events
		admin.POST("/create", app.HandleCreate)
		admin.PUT("/checkin", app.HandleCheckin)
		admin.PUT("/checkout", app.HandleCheckout)
		admin.PUT("/checkpoint", app.HandleCheckpoint)
		admin.POST("/add/volunteer", app.HandleVolunteers)
		admin.POST("/add/volunteerMan", app.HandleVolunteerManual)
	}

	if isHttpsServer == "1" {
		err := r.RunTLS(":8080", "localhost.crt", "localhost.key")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err := r.Run(":8080")
		if err != nil {
			log.Fatal(err)
		}
	}
}
