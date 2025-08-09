package handlers

import (
	"os"

	"github.com/gin-gonic/gin"
)

func CorsMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		if os.Getenv("PRODUCTION") == "true" {
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", os.Getenv("FRONTEND_ENDPOINT"))
		} else {
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", os.Getenv("FRONTEND_ENDPOINT"))
		}
		ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "content-type")

		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}
		ctx.Next()
	}
}
