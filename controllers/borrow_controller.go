package controllers

import (
	"OnlineLibraryPortal/database"
	"OnlineLibraryPortal/models"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func BorrowBook(c *gin.Context) {
	userRole := c.MustGet("userRole").(int)
	userName, _ := c.Get("userName")
	userEmail, _ := c.Get("userEmail")
	if userRole != 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only members can borrow books"})
		return
	}

	var req struct {
		BookID uint `json:"book_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID := c.MustGet("userID").(uint)

	var book models.Book
	if err := database.DB.First(&book, req.BookID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	if book.CopiesAvailable < 1 {
		c.JSON(http.StatusConflict, gin.H{"error": "No copies available"})
		return
	}

	borrowRecord := models.BorrowRecord{
		UserID:     userID,
		BookID:     req.BookID,
		BorrowedAt: time.Now(),
	}

	tx := database.DB.Begin()

	if err := tx.Create(&borrowRecord).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create borrow record"})
		return
	}

	book.CopiesAvailable -= 1
	if err := tx.Save(&book).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update book availability"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Book borrowed successfully"})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logEntry := map[string]interface{}{
		"user_id":    userID,
		"book_id":    req.BookID,
		"action":     "borrowed",
		"time":       time.Now(),
		"book_title": book.Title,
		"user_name":  userName,
		"email":      userEmail,
	}

	if _, err := database.BorrowingLogsCollection.InsertOne(ctx, logEntry); err != nil {
		log.Printf("Failed to insert borrowing log: %v", err)
	}
}

func ReturnBook(c *gin.Context) {
	userRole := c.MustGet("userRole").(int)
	userName, _ := c.Get("userName")
	userEmail, _ := c.Get("userEmail")
	if userRole != 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only members can return books"})
		return
	}

	var req struct {
		BookID uint `json:"book_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID := c.MustGet("userID").(uint)

	var borrowRecord models.BorrowRecord
	err := database.DB.
		Where("user_id = ? AND book_id = ? AND returned_at IS NULL", userID, req.BookID).
		Order("borrowed_at desc").
		First(&borrowRecord).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active borrow record found"})
		return
	}

	now := time.Now()
	borrowRecord.ReturnedAt = &now

	tx := database.DB.Begin()

	if err := tx.Save(&borrowRecord).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update borrow record"})
		return
	}

	if err := tx.Model(&models.Book{}).Where("id = ?", req.BookID).Update("copies_available", gorm.Expr("copies_available + ?", 1)).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update book availability"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Book returned successfully"})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var book models.Book
	if err := database.DB.First(&book, req.BookID).Error; err != nil {
		log.Printf("Failed to fetch book title for return log: %v", err)
	}

	logEntry := map[string]interface{}{
		"user_id":    userID,
		"book_id":    req.BookID,
		"book_title": book.Title,
		"action":     "returned",
		"time":       time.Now(),
		"user_name":  userName,
		"email":      userEmail,
	}

	if _, err := database.BorrowingLogsCollection.InsertOne(ctx, logEntry); err != nil {
		log.Printf("Failed to insert return log: %v", err)
	}
}

func BorrowingHistory(c *gin.Context) {
	fmt.Println("BorrowingHistory handler called")

	userID := c.MustGet("userID").(uint)
	userRole := c.MustGet("userRole").(int)

	fmt.Println("UserID:", userID)
	fmt.Println("UserRole:", userRole)

	var records []models.BorrowRecord
	var err error

	if userRole == 1 || userRole == 2 {
		fmt.Println("Role is librarian, fetching all records")

		err = database.DB.Preload("Book").
			Preload("User", func(db *gorm.DB) *gorm.DB {
				return db.Select("id", "name", "email", "jti", "role")
			}).
			Find(&records).Error
	} else {
		fmt.Println("Role is member, fetching their own history")
		err = database.DB.Preload("Book").
			Preload("User", func(db *gorm.DB) *gorm.DB {
				return db.Select("id", "name", "email", "jti", "role")
			}).
			Where("user_id = ?", userID).
			Find(&records).Error
	}

	if err != nil {
		fmt.Println("Error fetching borrowing history:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch borrowing history"})
		return
	}

	fmt.Println("Fetched borrow records count:", len(records))
	c.JSON(http.StatusOK, records)
}

func GetAllLibrarians(c *gin.Context) {
	var librarians []models.User

	if err := database.DB.
		Where("role = ?", 1).
		Select("id", "name", "email", "jti", "created_at", "updated_at").
		Find(&librarians).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch librarians"})
		return
	}

	c.JSON(http.StatusOK, librarians)
}

func GetAllMembers(c *gin.Context) {
	var members []models.User

	if err := database.DB.
		Where("role = ?", 0).
		Select("id", "name", "email", "jti", "created_at", "updated_at").
		Find(&members).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch members"})
		return
	}

	c.JSON(http.StatusOK, members)
}

func GetAdminDashboard(c *gin.Context) {
	type DashboardData struct {
		NumLibrarians        int64 `json:"num_librarians"`
		NumMembers           int64 `json:"num_members"`
		TotalCopies          int64 `json:"total_copies"`
		TotalCopiesAvailable int64 `json:"total_copies_available"`
	}

	var data DashboardData

	// Count librarians
	if err := database.DB.Model(&models.User{}).
		Where("role = ?", 1).
		Count(&data.NumLibrarians).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count librarians"})
		return
	}

	// Count members
	if err := database.DB.Model(&models.User{}).
		Where("role = ?", 0).
		Count(&data.NumMembers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count members"})
		return
	}

	// Sum total copies of books
	if err := database.DB.Model(&models.Book{}).
		Select("SUM(total_copies)").
		Scan(&data.TotalCopies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sum total copies"})
		return
	}

	// Sum available copies of books
	if err := database.DB.Model(&models.Book{}).
		Select("SUM(copies_available)").
		Scan(&data.TotalCopiesAvailable).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sum available copies"})
		return
	}

	c.JSON(http.StatusOK, data)
}

func GetOverdueBooks(c *gin.Context) {
	userRole := c.MustGet("userRole").(int)

	if userRole != 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only librarians can access overdue books"})
		return
	}

	cutoff := time.Now().AddDate(0, 0, 0)

	var overdueRecords []models.BorrowRecord

	err := database.DB.
		Preload("Book").
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "email", "role")
		}).
		Where("borrowed_at <= ? AND returned_at IS NULL", cutoff).
		Find(&overdueRecords).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch overdue books"})
		return
	}

	c.JSON(http.StatusOK, overdueRecords)
}
