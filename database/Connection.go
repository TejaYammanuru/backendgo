package database

import (
	"OnlineLibraryPortal/models"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {

	dsn := "root:12345@tcp(127.0.0.1:3306)/library_portal_development?charset=utf8mb4&parseTime=True&loc=Local"

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	fmt.Println("Successfully connected to the database")
	DB.AutoMigrate(&models.Book{}, &models.BorrowRecord{})
	DB.AutoMigrate(&models.BorrowRequest{})

}
