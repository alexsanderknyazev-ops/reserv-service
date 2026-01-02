package database

import (
	"log"
	"reserv-service/config"
	"reserv-service/models"

	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dbConfig := config.LoadDBConfig()

	var err error
	DB, err = config.ConnectDB(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")

	err = DB.AutoMigrate(&models.Reserve{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migrated successfully")
}

func GetDB() *gorm.DB {
	return DB
}
