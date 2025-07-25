package database

import (
	"fmt"
	"log"
	"os"

	"github.com/TLeTu/Chess-Media/server/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection established successfully.")
}

// UpdateUserELO updates a user's ELO in the database
func UpdateUserELO(userID uint, newELO int) error {
	result := DB.Model(&models.User{}).Where("id = ?", userID).Update("ELO", newELO)
	if result.Error != nil {
		return fmt.Errorf("failed to update ELO for user %d: %w", userID, result.Error)
	}
	return nil
}
