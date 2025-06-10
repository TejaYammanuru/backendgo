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

// func BorrowBook(c *gin.Context) {
// 	userRole := c.MustGet("userRole").(int)
// 	userName, _ := c.Get("userName")
// 	userEmail, _ := c.Get("userEmail")
// 	if userRole != 0 {
// 		c.JSON(http.StatusForbidden, gin.H{"error": "Only members can borrow books"})
// 		return
// 	}

// 	var req struct {
// 		BookID uint `json:"book_id"`
// 	}
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
// 		return
// 	}

// 	userID := c.MustGet("userID").(uint)

// 	var book models.Book
// 	if err := database.DB.First(&book, req.BookID).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
// 		return
// 	}

// 	if book.CopiesAvailable < 1 {
// 		c.JSON(http.StatusConflict, gin.H{"error": "No copies available"})
// 		return
// 	}

// 	borrowRecord := models.BorrowRecord{
// 		UserID:     userID,
// 		BookID:     req.BookID,
// 		BorrowedAt: time.Now(),
// 	}

// 	tx := database.DB.Begin()

// 	if err := tx.Create(&borrowRecord).Error; err != nil {
// 		tx.Rollback()
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create borrow record"})
// 		return
// 	}

// 	book.CopiesAvailable -= 1
// 	if err := tx.Save(&book).Error; err != nil {
// 		tx.Rollback()
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update book availability"})
// 		return
// 	}

// 	tx.Commit()
// 	c.JSON(http.StatusOK, gin.H{"message": "Book borrowed successfully"})

// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	logEntry := map[string]interface{}{
// 		"user_id":    userID,
// 		"book_id":    req.BookID,
// 		"action":     "borrowed",
// 		"time":       time.Now(),
// 		"book_title": book.Title,
// 		"user_name":  userName,
// 		"email":      userEmail,
// 	}

// 	if _, err := database.BorrowingLogsCollection.InsertOne(ctx, logEntry); err != nil {
// 		log.Printf("Failed to insert borrowing log: %v", err)
// 	}
// }

// func ReturnBook(c *gin.Context) {
// 	userRole := c.MustGet("userRole").(int)
// 	userName, _ := c.Get("userName")
// 	userEmail, _ := c.Get("userEmail")
// 	if userRole != 0 {
// 		c.JSON(http.StatusForbidden, gin.H{"error": "Only members can return books"})
// 		return
// 	}

// 	var req struct {
// 		BookID uint `json:"book_id"`
// 	}
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
// 		return
// 	}

// 	userID := c.MustGet("userID").(uint)

// 	var borrowRecord models.BorrowRecord
// 	err := database.DB.
// 		Where("user_id = ? AND book_id = ? AND returned_at IS NULL", userID, req.BookID).
// 		Order("borrowed_at desc").
// 		First(&borrowRecord).Error

// 	if err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "No active borrow record found"})
// 		return
// 	}

// 	now := time.Now()
// 	borrowRecord.ReturnedAt = &now

// 	tx := database.DB.Begin()

// 	if err := tx.Save(&borrowRecord).Error; err != nil {
// 		tx.Rollback()
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update borrow record"})
// 		return
// 	}

// 	if err := tx.Model(&models.Book{}).Where("id = ?", req.BookID).Update("copies_available", gorm.Expr("copies_available + ?", 1)).Error; err != nil {
// 		tx.Rollback()
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update book availability"})
// 		return
// 	}

// 	tx.Commit()
// 	c.JSON(http.StatusOK, gin.H{"message": "Book returned successfully"})

// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	var book models.Book
// 	if err := database.DB.First(&book, req.BookID).Error; err != nil {
// 		log.Printf("Failed to fetch book title for return log: %v", err)
// 	}

// 	logEntry := map[string]interface{}{
// 		"user_id":    userID,
// 		"book_id":    req.BookID,
// 		"book_title": book.Title,
// 		"action":     "returned",
// 		"time":       time.Now(),
// 		"user_name":  userName,
// 		"email":      userEmail,
// 	}

// 	if _, err := database.BorrowingLogsCollection.InsertOne(ctx, logEntry); err != nil {
// 		log.Printf("Failed to insert return log: %v", err)
// 	}
// }

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

func BorrowRequest(c *gin.Context) {
	log.Println("📥 BorrowRequest handler triggered")

	userRole := c.MustGet("userRole").(int)
	userID := c.MustGet("userID").(uint)
	log.Printf("🔐 UserID: %d, Role: %d\n", userID, userRole)

	if userRole != 0 {
		log.Println("❌ Access denied: Only members can request books")
		c.JSON(http.StatusForbidden, gin.H{"error": "Only members can request books"})
		return
	}

	var req struct {
		BookID uint `json:"book_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Failed to parse JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	log.Printf("📚 Book request received for BookID: %d\n", req.BookID)

	// Check if book exists
	var book models.Book
	if err := database.DB.First(&book, req.BookID).Error; err != nil {
		log.Printf("❌ Book not found with ID %d: %v\n", req.BookID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}
	log.Printf("✅ Book found: %s\n", book.Title)

	// Check for existing pending request
	var existingRequest models.BorrowRequest
	err := database.DB.Where("user_id = ? AND book_id = ? AND status = ?", userID, req.BookID, "pending").
		First(&existingRequest).Error
	if err == nil {
		log.Println("⚠️ Duplicate pending request exists")
		c.JSON(http.StatusConflict, gin.H{"error": "You already have a pending request for this book"})
		return
	}
	if err != gorm.ErrRecordNotFound {
		log.Printf("❌ Error checking existing request: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	// Create request
	request := models.BorrowRequest{
		UserID:      userID,
		BookID:      req.BookID,
		Status:      "pending",
		RequestedAt: time.Now(),
	}

	if err := database.DB.Create(&request).Error; err != nil {
		log.Printf("❌ Failed to create borrow request: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	log.Println("✅ Borrow request created successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Book request submitted successfully"})
}

func GetAllBorrowRequests(c *gin.Context) {
	log.Println("GetAllBorrowRequests handler triggered")

	userRole := c.MustGet("userRole").(int)
	if userRole != 1 {
		log.Println("Access denied: Only librarians can view borrow requests")
		c.JSON(http.StatusForbidden, gin.H{"error": "Only librarians can view borrow requests"})
		return
	}

	var requests []models.BorrowRequest
	if err := database.DB.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "email", "role")
		}).
		Preload("Book").
		Order("requested_at desc").
		Find(&requests).Error; err != nil {
		log.Printf(" Failed to fetch borrow requests: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch requests"})
		return
	}

	log.Printf("%d borrow requests fetched", len(requests))
	c.JSON(http.StatusOK, requests)
}

func ApproveBorrowRequest(c *gin.Context) {
	log.Println("✅ ApproveBorrowRequest handler triggered")

	userRole := c.MustGet("userRole").(int)
	librarianName, _ := c.Get("userName")
	librarianEmail, _ := c.Get("userEmail")

	if userRole != 1 {
		log.Println("❌ Access denied: Only librarians can approve borrow requests")
		c.JSON(http.StatusForbidden, gin.H{"error": "Only librarians can approve borrow requests"})
		return
	}

	var req struct {
		RequestID uint `json:"request_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.RequestID == 0 {
		log.Printf("❌ Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var borrowRequest models.BorrowRequest
	if err := database.DB.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "email")
		}).
		Preload("Book").
		First(&borrowRequest, req.RequestID).Error; err != nil {
		log.Printf("❌ Borrow request not found: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Borrow request not found"})
		return
	}

	if borrowRequest.Status != "pending" {
		log.Println("⚠️ Borrow request is not pending")
		c.JSON(http.StatusConflict, gin.H{"error": "Request is not pending"})
		return
	}

	if borrowRequest.Book.CopiesAvailable < 1 {
		log.Println("❌ No available copies to approve this request")
		c.JSON(http.StatusConflict, gin.H{"error": "No available copies of the book"})
		return
	}

	tx := database.DB.Begin()

	// Create borrow record
	borrowRecord := models.BorrowRecord{
		UserID:     borrowRequest.UserID,
		BookID:     borrowRequest.BookID,
		BorrowedAt: time.Now(),
	}
	if err := tx.Create(&borrowRecord).Error; err != nil {
		tx.Rollback()
		log.Printf("❌ Failed to create borrow record: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create borrow record"})
		return
	}

	// Decrease book copies
	if err := tx.Model(&models.Book{}).
		Where("id = ?", borrowRequest.BookID).
		Update("copies_available", gorm.Expr("copies_available - 1")).Error; err != nil {
		tx.Rollback()
		log.Printf("❌ Failed to update book availability: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update book availability"})
		return
	}

	// Update only relevant fields
	now := time.Now()
	if err := tx.Model(&models.BorrowRequest{}).
		Where("id = ?", req.RequestID).
		Updates(map[string]interface{}{
			"status":      "approved",
			"approved_at": now,
		}).Error; err != nil {
		tx.Rollback()
		log.Printf("❌ Failed to update borrow request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update request status"})
		return
	}

	tx.Commit()

	// MongoDB log
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logEntry := map[string]interface{}{
		"action":         "request_approved",
		"book_title":     borrowRequest.Book.Title,
		"user_id":        borrowRequest.User.ID,
		"user_name":      borrowRequest.User.Name,
		"user_email":     borrowRequest.User.Email,
		"approved_by":    librarianName,
		"approver_email": librarianEmail,
		"time":           now,
	}
	if _, err := database.BorrowingLogsCollection.InsertOne(ctx, logEntry); err != nil {
		log.Printf("⚠️ Failed to insert log: %v", err)
	}

	log.Println("✅ Request approved successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Borrow request approved and borrow record created"})
}

func RejectBorrowRequest(c *gin.Context) {
	log.Println("❌ RejectBorrowRequest handler triggered")

	userRole := c.MustGet("userRole").(int)
	librarianName, _ := c.Get("userName")
	librarianEmail, _ := c.Get("userEmail")

	if userRole != 1 {
		log.Println("❌ Access denied: Only librarians can reject borrow requests")
		c.JSON(http.StatusForbidden, gin.H{"error": "Only librarians can reject borrow requests"})
		return
	}

	var req struct {
		RequestID uint   `json:"request_id"`
		Reason    string `json:"reason"` // ✅ Added reason
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.RequestID == 0 {
		log.Printf("❌ Invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var borrowRequest models.BorrowRequest
	if err := database.DB.First(&borrowRequest, req.RequestID).Error; err != nil {
		log.Printf("❌ Borrow request not found: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Borrow request not found"})
		return
	}

	if borrowRequest.Status != "pending" {
		log.Println("⚠️ Borrow request is not pending")
		c.JSON(http.StatusConflict, gin.H{"error": "Only pending requests can be rejected"})
		return
	}

	now := time.Now()
	reason := req.Reason // capture locally

	if err := database.DB.Model(&borrowRequest).Updates(map[string]interface{}{
		"status":           "rejected",
		"rejected_at":      now,
		"rejection_reason": reason,
	}).Error; err != nil {
		log.Printf("❌ Failed to update borrow request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update request"})
		return
	}

	// Fetch user and book for logs
	var user models.User
	_ = database.DB.Select("id", "name", "email").First(&user, borrowRequest.UserID)
	var book models.Book
	_ = database.DB.Select("id", "title").First(&book, borrowRequest.BookID)

	// MongoDB log
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logEntry := map[string]interface{}{
		"action":           "request_rejected",
		"book_title":       book.Title,
		"user_id":          user.ID,
		"user_name":        user.Name,
		"user_email":       user.Email,
		"rejected_by":      librarianName,
		"rejector_email":   librarianEmail,
		"rejection_reason": reason,
		"time":             now,
	}
	if _, err := database.BorrowingLogsCollection.InsertOne(ctx, logEntry); err != nil {
		log.Printf("⚠️ Failed to insert Mongo log: %v", err)
	}

	log.Println("✅ Borrow request rejected successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Borrow request rejected successfully"})
}

func ReturnRequest(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	userRole := c.MustGet("userRole").(int)

	if userRole != 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only members can request returns"})
		return
	}

	var req struct {
		BorrowID uint `json:"borrow_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.BorrowID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid borrow ID"})
		return
	}

	var record models.BorrowRecord
	if err := database.DB.First(&record, req.BorrowID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Borrow record not found"})
		return
	}

	// Check if record belongs to the requesting member
	if record.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access to this borrow record"})
		return
	}

	if record.ReturnedAt != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Book already returned"})
		return
	}

	if record.ReturnRequested {
		c.JSON(http.StatusConflict, gin.H{"error": "Return already requested"})
		return
	}

	record.ReturnRequested = true
	if err := database.DB.Save(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not request return"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Return request submitted"})
}

func AcknowledgeReturn(c *gin.Context) {
	userRole := c.MustGet("userRole").(int)

	if userRole != 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only librarians can acknowledge returns"})
		return
	}

	var req struct {
		BorrowID uint `json:"borrow_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.BorrowID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var record models.BorrowRecord
	if err := database.DB.First(&record, req.BorrowID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Borrow record not found"})
		return
	}

	if record.ReturnedAt != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Book already returned"})
		return
	}

	if !record.ReturnRequested {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Return not requested"})
		return
	}

	now := time.Now()
	record.ReturnedAt = &now
	record.ReturnRequested = false

	tx := database.DB.Begin()

	if err := tx.Save(&record).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to acknowledge return"})
		return
	}

	if err := tx.Model(&models.Book{}).
		Where("id = ?", record.BookID).
		Update("copies_available", gorm.Expr("copies_available + 1")).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update book availability"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Book return acknowledged"})
}
