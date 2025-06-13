package controllers

import (
	"OnlineLibraryPortal/database"
	"OnlineLibraryPortal/models"
	"OnlineLibraryPortal/utils"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
	userID := c.MustGet("userID").(uint)

	var records []models.BorrowRecord
	query := database.DB.
		Preload("Book").
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "email", "role")
		}).
		Where("returned_at IS NULL")

	if userRole == 0 {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch records"})
		return
	}

	type OverdueInfo struct {
		BorrowID       uint        `json:"borrow_id"`
		User           models.User `json:"user"`
		Book           models.Book `json:"book"`
		BorrowedAt     time.Time   `json:"borrowed_at"`
		ExpectedReturn time.Time   `json:"expected_return"`
		DaysOverdue    int         `json:"days_overdue"`
	}

	now := time.Now()
	result := make([]OverdueInfo, 0)

	for _, record := range records {
		expectedReturn := record.BorrowedAt.AddDate(0, 0, record.Book.OverdueDays)

		// ✅ Include only if overdue
		if now.After(expectedReturn) {
			daysOverdue := int(now.Sub(expectedReturn).Hours() / 24)

			result = append(result, OverdueInfo{
				BorrowID:       record.ID,
				User:           record.User,
				Book:           record.Book,
				BorrowedAt:     record.BorrowedAt,
				ExpectedReturn: expectedReturn,
				DaysOverdue:    daysOverdue,
			})
		}
	}

	c.JSON(http.StatusOK, result)
}

func BorrowRequest(c *gin.Context) {
	log.Println("📥 BorrowRequest handler triggered")

	userRole := c.MustGet("userRole").(int)
	userID := c.MustGet("userID").(uint)
	userName := c.MustGet("userName").(string)
	userEmail := c.MustGet("userEmail").(string)
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

	var book models.Book
	if err := database.DB.First(&book, req.BookID).Error; err != nil {
		log.Printf("Book not found with ID %d: %v\n", req.BookID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}
	log.Printf("Book found: %s\n", book.Title)

	var activeBorrow models.BorrowRecord
	err := database.DB.Where("user_id = ? AND book_id = ? AND returned_at IS NULL", userID, req.BookID).
		First(&activeBorrow).Error
	if err == nil {
		log.Println("Book already borrowed and not yet returned")
		c.JSON(http.StatusConflict, gin.H{"error": "You have already borrowed this book and haven't returned it yet"})
		return
	}
	if err != gorm.ErrRecordNotFound {
		log.Printf("Error checking active borrow: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	var existingRequest models.BorrowRequest
	err = database.DB.Where("user_id = ? AND book_id = ? AND status = ?", userID, req.BookID, "pending").
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

	// ✅ Create new borrow request
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

	// ✅ Log to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logEntry := map[string]interface{}{
		"action":     "borrow_request_created",
		"user_id":    userID,
		"user_name":  userName,
		"user_email": userEmail,
		"book_id":    req.BookID,
		"book_title": book.Title,
		"time":       time.Now(),
	}
	if _, err := database.BorrowingLogsCollection.InsertOne(ctx, logEntry); err != nil {
		log.Printf("⚠️ Failed to insert Mongo log: %v", err)
	}

	log.Println("✅ Borrow request created successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Book request submitted successfully"})
}

func GetAllBorrowRequests(c *gin.Context) {
	log.Println("📄 GetAllBorrowRequests handler triggered")

	userRole := c.MustGet("userRole").(int)
	if userRole != 1 {
		log.Println("❌ Access denied: Only librarians can view borrow requests")
		c.JSON(http.StatusForbidden, gin.H{"error": "Only librarians can view borrow requests"})
		return
	}

	var requests []models.BorrowRequest
	if err := database.DB.
		Where("status = ?", "pending"). // ✅ filter only pending requests
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "email", "role")
		}).
		Preload("Book").
		Order("requested_at desc").
		Find(&requests).Error; err != nil {
		log.Printf("❌ Failed to fetch borrow requests: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch requests"})
		return
	}

	log.Printf("✅ %d pending borrow requests fetched", len(requests))
	c.JSON(http.StatusOK, requests)
}

func ApproveBorrowRequest(c *gin.Context) {
	log.Println("✅ ApproveBorrowRequest handler triggered")

	userRole := c.MustGet("userRole").(int)
	librarianName := c.MustGet("userName").(string)
	librarianEmail := c.MustGet("userEmail").(string)

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

	// Update borrow request
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

	go utils.SendEmail(
		borrowRequest.User.Email,
		"Borrow Request Approved",
		fmt.Sprintf("Hi %s,\n\nYour request to borrow '%s' has been approved.\n\nHappy Reading!\n\n- Library Team",
			borrowRequest.User.Name, borrowRequest.Book.Title),
	)

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
	librarianName := c.MustGet("userName").(string)
	librarianEmail := c.MustGet("userEmail").(string)

	if userRole != 1 {
		log.Println("❌ Access denied: Only librarians can reject borrow requests")
		c.JSON(http.StatusForbidden, gin.H{"error": "Only librarians can reject borrow requests"})
		return
	}

	var req struct {
		RequestID uint   `json:"request_id"`
		Reason    string `json:"reason"`
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

	if err := database.DB.Model(&borrowRequest).Updates(map[string]interface{}{
		"status":           "rejected",
		"rejected_at":      now,
		"rejection_reason": req.Reason,
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
		"rejection_reason": req.Reason,
		"time":             now,
	}
	if _, err := database.BorrowingLogsCollection.InsertOne(ctx, logEntry); err != nil {
		log.Printf("⚠️ Failed to insert Mongo log: %v", err)
	}

	log.Println("✅ Borrow request rejected successfully")

	go utils.SendEmail(
		user.Email,
		"Borrow Request Rejected",
		fmt.Sprintf("Hi %s,\n\nYour borrow request for '%s' has been rejected.\nReason: %s\n\nRegards,\nLibrary Team",
			user.Name, book.Title, req.Reason),
	)

	c.JSON(http.StatusOK, gin.H{"message": "Borrow request rejected successfully"})
}

func ReturnRequest(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	userRole := c.MustGet("userRole").(int)
	userName := c.MustGet("userName").(string)
	userEmail := c.MustGet("userEmail").(string)

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

	// MongoDB log
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logEntry := map[string]interface{}{
		"action":     "return_requested",
		"user_id":    userID,
		"user_name":  userName,
		"user_email": userEmail,
		"borrow_id":  req.BorrowID,
		"time":       time.Now(),
	}
	if _, err := database.MongoClient.Database("library_portal_logging").
		Collection("return_logs").InsertOne(ctx, logEntry); err != nil {
		log.Printf("⚠️ Failed to insert return log: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Return request submitted"})
}

func AcknowledgeReturn(c *gin.Context) {
	userRole := c.MustGet("userRole").(int)
	librarianName := c.MustGet("userName").(string)
	librarianEmail := c.MustGet("userEmail").(string)

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

	var book models.Book
	if err := database.DB.Select("id", "title").First(&book, record.BookID).Error; err != nil {
		log.Printf("Failed to fetch book for ID %d: %v", record.BookID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch book details"})
		return
	}

	var user models.User
	if err := database.DB.Select("id", "name", "email").First(&user, record.UserID).Error; err == nil {

		go utils.SendEmail(
			user.Email,
			"Book Return Acknowledged",
			fmt.Sprintf("Hi %s,\n\nYour returned book %s has been acknowledged successfully. Thank you!\n\n- Library Team",
				user.Name, book.Title),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		logEntry := map[string]interface{}{
			"action":          "return_acknowledged",
			"user_id":         user.ID,
			"user_name":       user.Name,
			"user_email":      user.Email,
			"borrow_id":       req.BorrowID,
			"acknowledged_by": librarianName,
			"ack_email":       librarianEmail,
			"time":            now,
		}
		if _, err := database.MongoClient.Database("library_portal_logging").
			Collection("return_logs").InsertOne(ctx, logEntry); err != nil {
			log.Printf("Failed to insert return log: %v", err)
		}
	} else {
		log.Printf("Failed to fetch user for email: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Book return acknowledged"})
}

func GetMyBorrowRequests(c *gin.Context) {
	log.Println("📩 GetMyBorrowRequests handler triggered")

	userID := c.MustGet("userID").(uint)

	var requests []models.BorrowRequest
	err := database.DB.
		Where("user_id = ?", userID).
		Preload("Book").
		Order("requested_at desc").
		Find(&requests).Error

	if err != nil {
		log.Printf("❌ Failed to fetch user borrow requests: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch borrow requests"})
		return
	}

	log.Printf("✅ %d borrow requests fetched for user %d", len(requests), userID)
	c.JSON(http.StatusOK, requests)
}

func GetBooksNotYetReturned(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var borrowRecords []models.BorrowRecord

	err := database.DB.
		Where("user_id = ? AND returned_at IS NULL", userID).
		Preload("Book").
		Order("borrowed_at desc").
		Find(&borrowRecords).Error

	if err != nil {
		log.Printf("❌ Failed to fetch borrowed books: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch active borrow records"})
		return
	}

	type BookWithDueInfo struct {
		models.BorrowRecord
		DaysLeft int `json:"days_left"`
	}

	var result []BookWithDueInfo
	now := time.Now()

	for _, record := range borrowRecords {
		dueDays := record.Book.OverdueDays
		daysPassed := int(now.Sub(record.BorrowedAt).Hours() / 24)
		daysLeft := dueDays - daysPassed
		if daysLeft < 0 {
			daysLeft = 0
		}

		result = append(result, BookWithDueInfo{
			BorrowRecord: record,
			DaysLeft:     daysLeft,
		})
	}

	c.JSON(http.StatusOK, result)
}

func GetBooksReturnRequestedNotAcknowledged(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var records []models.BorrowRecord

	err := database.DB.
		Where("user_id = ? AND return_requested = ? AND returned_at IS NULL", userID, true).
		Preload("Book").
		Order("borrowed_at desc").
		Find(&records).Error

	if err != nil {
		log.Printf("❌ Failed to fetch return requested books: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch return requested books"})
		return
	}

	c.JSON(http.StatusOK, records)
}

func GetReturnPendingRecords(c *gin.Context) {
	log.Println("📄 GetReturnPendingRecords handler triggered")

	userRole := c.MustGet("userRole").(int)
	if userRole != 1 {
		log.Println("❌ Access denied: Only librarians can view return requests")
		c.JSON(http.StatusForbidden, gin.H{"error": "Only librarians can access this resource"})
		return
	}

	var records []models.BorrowRecord
	err := database.DB.
		Where("return_requested = ? AND returned_at IS NULL", true).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "email", "role")
		}).
		Preload("Book").
		Order("borrowed_at desc").
		Find(&records).Error

	if err != nil {
		log.Printf("❌ Failed to fetch return-pending records: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch return requests"})
		return
	}

	log.Printf("✅ %d return-pending records fetched", len(records))
	c.JSON(http.StatusOK, records)
}

func GetLibrarianDashboardStats(c *gin.Context) {
	userRole := c.MustGet("userRole").(int)
	if userRole != 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only librarians can access dashboard"})
		return
	}

	type DashboardStats struct {
		TotalBooks         int64 `json:"total_books"`
		AvailableBooks     int64 `json:"available_books"`
		TotalMembers       int64 `json:"total_members"`
		TotalLibrarians    int64 `json:"total_librarians"`
		PendingBorrowCount int64 `json:"pending_borrow_count"`
		PendingReturnCount int64 `json:"pending_return_count"`
	}

	var stats DashboardStats

	// Total books
	if err := database.DB.Model(&models.Book{}).Select("SUM(total_copies)").Scan(&stats.TotalBooks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count total books"})
		return
	}

	// Available books
	if err := database.DB.Model(&models.Book{}).Select("SUM(copies_available)").Scan(&stats.AvailableBooks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count available books"})
		return
	}

	// Total members
	if err := database.DB.Model(&models.User{}).Where("role = ?", 0).Count(&stats.TotalMembers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count members"})
		return
	}

	// Total librarians
	if err := database.DB.Model(&models.User{}).Where("role = ?", 1).Count(&stats.TotalLibrarians).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count librarians"})
		return
	}

	// Pending borrow requests
	if err := database.DB.Model(&models.BorrowRequest{}).Where("status = ?", "pending").Count(&stats.PendingBorrowCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count pending borrow requests"})
		return
	}

	// Pending return acknowledgments
	if err := database.DB.Model(&models.BorrowRecord{}).
		Where("return_requested = ? AND returned_at IS NULL", true).
		Count(&stats.PendingReturnCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count pending returns"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func GetMemberNotifications(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	userRole := c.MustGet("userRole").(int)

	if userRole != 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only members can view notifications"})
		return
	}

	now := time.Now()
	fiveDaysAgo := now.AddDate(0, 0, -5)
	twoDaysAgo := now.AddDate(0, 0, -2)
	oneMonthAgo := now.AddDate(0, -1, 0)

	var newBooks, updatedBooks, deletedBooks []models.Book
	var dueSoon []models.BorrowRecord

	db := database.DB

	db.Where("created_at >= ?", twoDaysAgo).Find(&newBooks)

	db.Where("updated_at >= ? AND updated_at != created_at", twoDaysAgo).Find(&updatedBooks)

	db.Unscoped().Where("deleted_at IS NOT NULL AND deleted_at >= ?", fiveDaysAgo).Find(&deletedBooks)

	db.Preload("Book").Where("user_id = ? AND returned_at IS NULL", userID).
		Find(&dueSoon)
	dueNotifications := []string{}
	for _, record := range dueSoon {
		dueDate := record.BorrowedAt.AddDate(0, 0, record.Book.OverdueDays)
		if dueDate.After(now) && dueDate.Before(now.AddDate(0, 0, 4)) {
			daysLeft := int(dueDate.Sub(now).Hours() / 24)
			dueNotifications = append(dueNotifications,
				fmt.Sprintf("'%s' is due in %d day(s). Please return on time.", record.Book.Title, daysLeft))
		}
	}

	type BorrowCount struct {
		BookID uint
		Title  string
		Count  int
	}
	var popular []BorrowCount
	db.Table("borrow_records").
		Select("book_id, books.title, COUNT(*) as count").
		Joins("JOIN books ON books.id = borrow_records.book_id").
		Where("borrowed_at >= ?", oneMonthAgo).
		Group("book_id, books.title").
		Order("count DESC").
		Limit(5).
		Scan(&popular)

	messages := []string{}
	for _, book := range newBooks {
		messages = append(messages, fmt.Sprintf("New Book Added: '%s'", book.Title))
	}
	for _, book := range updatedBooks {
		messages = append(messages, fmt.Sprintf("Book Updated: '%s'", book.Title))
	}
	for _, book := range deletedBooks {
		messages = append(messages, fmt.Sprintf("Book Removed: '%s'", book.Title))
	}
	messages = append(messages, dueNotifications...)
	for _, pop := range popular {
		messages = append(messages, fmt.Sprintf("Popular Book: '%s' borrowed %d time(s) last month", pop.Title, pop.Count))
	}

	c.JSON(http.StatusOK, gin.H{"notifications": messages})
}

func GetMemberOverview(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	userRole := c.MustGet("userRole").(int)

	if userRole != 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only members can view this overview"})
		return
	}

	db := database.DB

	var totalBorrowed int64
	var currentlyBorrowed int64
	var overdueCount int
	var totalReturnRequests int64
	var acknowledgedReturns int64
	var totalBooks int64

	// 1. Total borrowed by user
	db.Model(&models.BorrowRecord{}).Where("user_id = ?", userID).Count(&totalBorrowed)

	// 2. Currently borrowed (not returned)
	db.Model(&models.BorrowRecord{}).
		Where("user_id = ? AND returned_at IS NULL", userID).
		Count(&currentlyBorrowed)

	// 3. Overdue books
	var records []models.BorrowRecord
	db.Preload("Book").Where("user_id = ? AND returned_at IS NULL", userID).Find(&records)
	now := time.Now()
	for _, rec := range records {
		due := rec.BorrowedAt.AddDate(0, 0, rec.Book.OverdueDays)
		if now.After(due) {
			overdueCount++
		}
	}

	// 4. Return requests made
	db.Model(&models.BorrowRecord{}).
		Where("user_id = ? AND return_requested = true", userID).
		Count(&totalReturnRequests)

	// 5. Acknowledged returns
	db.Model(&models.BorrowRecord{}).
		Where("user_id = ? AND returned_at IS NOT NULL", userID).
		Count(&acknowledgedReturns)

	// 6. Total books in library
	db.Model(&models.Book{}).Count(&totalBooks)

	// 7. Recent borrow records
	var recentRecords []models.BorrowRecord
	db.Preload("Book").
		Where("user_id = ?", userID).
		Order("borrowed_at DESC").
		Limit(5).
		Find(&recentRecords)

	type Recent struct {
		Title      string    `json:"title"`
		BorrowedAt time.Time `json:"borrowed_at"`
	}

	recent := []Recent{}
	for _, r := range recentRecords {
		recent = append(recent, Recent{
			Title:      r.Book.Title,
			BorrowedAt: r.BorrowedAt,
		})
	}

	// 🔄 Final response
	c.JSON(http.StatusOK, gin.H{
		"total_borrowed":        totalBorrowed,
		"currently_borrowed":    currentlyBorrowed,
		"overdue_books":         overdueCount,
		"return_requested":      totalReturnRequests,
		"acknowledged_returns":  acknowledgedReturns,
		"total_books_available": totalBooks,
		"recent_borrows":        recent,
	})
}
