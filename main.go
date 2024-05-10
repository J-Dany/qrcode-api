package main

import (
	"context"
	"os"
	"qrcode/api"
	"qrcode/database"
	"qrcode/env"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load .env file
	env.LoadEnvs()

	// Create a new router
	router := gin.Default()

	router.Use(gin.LoggerWithWriter(gin.DefaultWriter, "qrcode"))

	// Setup API
	api.SetupAPI(router)

	// Run the server
	router.Run("0.0.0.0:" + os.Getenv("PORT"))

	// Close the database connection
	defer database.Client.Disconnect(context.Background())
}
