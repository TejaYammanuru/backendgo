package main

import (
	"OnlineLibraryPortal/database"
	"OnlineLibraryPortal/routes"

	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	database.Connect()
	database.ConnectMongo()

	r := gin.Default()
	r.RedirectTrailingSlash = false

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Static("/uploads", "./uploads")

	routes.BookRoutes(r)
	routes.BorrowRoutes(r)

	r.Run(":8080")
}
