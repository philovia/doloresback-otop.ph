package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/joho/godotenv"

	"github.com/m/models"
)

var DB *gorm.DB

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func SetupDatabase() {
	LoadEnv()

	dsn := "host=localhost port=5432 user=postgres dbname=postgres password=postgres"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	log.Println("Successfully connected to PostgreSQL using GORM")

	DB.AutoMigrate(&models.Product{}, &models.User{}, &models.Supplier{}, &models.OtopProducts{},
		&models.SoldItems{}, &models.Transaction{}, &models.TransactionItem{}, &models.TransactionSupplier{})
}
