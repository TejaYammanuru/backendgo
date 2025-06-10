package controllers

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"OnlineLibraryPortal/database"
	"OnlineLibraryPortal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func isAuthorizedToModify(role int) bool {
	return role == 1 || role == 2
}

type BookRequest struct {
	Title           string `json:"title" binding:"required"`
	Author          string `json:"author" binding:"required"`
	Genre           string `json:"genre" binding:"required"`
	PublicationDate string `json:"publication_date"`
	TotalCopies     int    `json:"total_copies"`
	CopiesAvailable int    `json:"copies_available"`
	ImageURL        string `json:"image_url"`
}

func CreateBook(c *gin.Context) {
	userRole := c.GetInt("userRole")
	userName, _ := c.Get("userName")
	userEmail, _ := c.Get("userEmail")

	if !isAuthorizedToModify(userRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	var req BookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	if req.TotalCopies < 0 || req.CopiesAvailable < 0 || req.CopiesAvailable > req.TotalCopies {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid copies values"})
		return
	}

	var pubDate time.Time
	var err error
	if req.PublicationDate != "" {
		dateFormats := []string{
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			"2006-01-02T15:04Z",
			"2006-01-02",
		}
		for _, format := range dateFormats {
			pubDate, err = time.Parse(format, req.PublicationDate)
			if err == nil {
				break
			}
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid publication_date format"})
			return
		}
	} else {
		pubDate = time.Now()
	}

	imageUrl := ""
	if req.ImageURL != "" {
		if strings.HasPrefix(req.ImageURL, "data:image/") {
			imageUrl, err = saveBase64Image(req.ImageURL)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image: " + err.Error()})
				return
			}
		} else {
			imageUrl = req.ImageURL
		}
	}

	book := models.Book{
		Title:           req.Title,
		Author:          req.Author,
		PublicationDate: pubDate,
		Genre:           req.Genre,
		TotalCopies:     req.TotalCopies,
		CopiesAvailable: req.CopiesAvailable,
		ImageURL:        imageUrl,
	}

	if err := database.DB.Create(&book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create book"})
		return
	}

	go func() {
		ctx := context.Background()
		logEntry := bson.M{
			"operation":       "create",
			"book_id":         book.ID,
			"title":           book.Title,
			"author":          book.Author,
			"genre":           book.Genre,
			"publicationDate": book.PublicationDate,
			"totalCopies":     book.TotalCopies,
			"copiesAvailable": book.CopiesAvailable,
			"imageURL":        book.ImageURL,
			"userRole":        userRole,
			"userName":        userName,
			"userEmail":       userEmail,
			"timestamp":       time.Now(),
		}
		_, err := database.BookLogsCollection.InsertOne(ctx, logEntry)
		if err != nil {
			fmt.Println("Failed to insert book create log:", err)
		}
	}()

	c.JSON(http.StatusCreated, book)
}

func UpdateBook(c *gin.Context) {
	userRole := c.GetInt("userRole")
	userName, _ := c.Get("userName")
	userEmail, _ := c.Get("userEmail")

	if !isAuthorizedToModify(userRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	bookID := c.Param("id")
	var req BookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	if req.TotalCopies < 0 || req.CopiesAvailable < 0 || req.CopiesAvailable > req.TotalCopies {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid copies values"})
		return
	}

	var pubDate time.Time
	var err error
	if req.PublicationDate != "" {
		dateFormats := []string{
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			"2006-01-02T15:04Z",
			"2006-01-02",
		}
		for _, format := range dateFormats {
			pubDate, err = time.Parse(format, req.PublicationDate)
			if err == nil {
				break
			}
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid publication_date format"})
			return
		}
	}

	var book models.Book
	if err := database.DB.First(&book, bookID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	prevBook := book

	imageUrl := book.ImageURL
	if req.ImageURL != "" {
		if strings.HasPrefix(req.ImageURL, "data:image/") {
			imageUrl, err = saveBase64Image(req.ImageURL)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image: " + err.Error()})
				return
			}
		} else if req.ImageURL != book.ImageURL {
			imageUrl = req.ImageURL
		}
	}

	book.Title = req.Title
	book.Author = req.Author
	book.Genre = req.Genre
	book.PublicationDate = pubDate
	book.TotalCopies = req.TotalCopies
	book.CopiesAvailable = req.CopiesAvailable
	book.ImageURL = imageUrl

	if err := database.DB.Save(&book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update book"})
		return
	}

	go func() {
		ctx := context.Background()
		logEntry := bson.M{
			"operation": "update",
			"book_id":   book.ID,
			"userRole":  userRole,
			"userName":  userName,
			"userEmail": userEmail,
			"timestamp": time.Now(),
			"before":    prevBook,
			"after":     book,
		}
		_, err := database.BookLogsCollection.InsertOne(ctx, logEntry)
		if err != nil {
			fmt.Println("Failed to insert book update log:", err)
		}
	}()

	c.JSON(http.StatusOK, book)
}

func DeleteBook(c *gin.Context) {
	userRole := c.GetInt("userRole")
	userName, _ := c.Get("userName")
	userEmail, _ := c.Get("userEmail")

	if !isAuthorizedToModify(userRole) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	id := c.Param("id")
	var book models.Book
	if err := database.DB.First(&book, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	if err := database.DB.Delete(&book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete book"})
		return
	}

	go func() {
		ctx := context.Background()
		logEntry := bson.M{
			"operation": "delete",
			"book_id":   book.ID,
			"title":     book.Title,
			"userRole":  userRole,
			"userName":  userName,
			"userEmail": userEmail,
			"timestamp": time.Now(),
		}
		_, err := database.BookLogsCollection.InsertOne(ctx, logEntry)
		if err != nil {
			fmt.Println("Failed to insert book delete log:", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Book deleted"})
}

func saveBase64Image(base64String string) (string, error) {

	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		return "", err
	}

	parts := strings.Split(base64String, ",")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid base64 format")
	}

	var ext string
	if strings.Contains(parts[0], "jpeg") {
		ext = ".jpg"
	} else if strings.Contains(parts[0], "png") {
		ext = ".png"
	} else if strings.Contains(parts[0], "gif") {
		ext = ".gif"
	} else {
		ext = ".jpg"
	}

	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	filename := strconv.FormatInt(time.Now().UnixNano(), 10) + ext
	filepath := "uploads/" + filename

	file, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return "", err
	}

	return "/" + filepath, nil
}

func GetBooks(c *gin.Context) {
	var books []models.Book
	database.DB.Find(&books)
	c.JSON(http.StatusOK, books)
}

func GetBook(c *gin.Context) {
	id := c.Param("id")
	var book models.Book
	if err := database.DB.First(&book, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}
	c.JSON(http.StatusOK, book)
}
