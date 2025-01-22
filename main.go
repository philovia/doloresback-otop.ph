package main

import (
	// "fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/m/database"
	"github.com/m/models"
	"github.com/m/routes"
)

func main() {

	database.SetupDatabase()

	if err := database.DB.AutoMigrate(&models.Supplier{}); err != nil {
		log.Fatalf("Could not auto-migrate supplier table: %v", err)
	}

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://philovia.github.io",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	routes.UserRoutes(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8097"
	}

	log.Fatal(app.Listen(":" + port))
}
