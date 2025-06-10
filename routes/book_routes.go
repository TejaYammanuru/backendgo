package routes

import (
	"OnlineLibraryPortal/controllers"
	"OnlineLibraryPortal/middleware"

	"github.com/gin-gonic/gin"
)

func BookRoutes(router *gin.Engine) {
	books := router.Group("/books")

	books.GET("/", controllers.GetBooks)
	books.GET("/:id", controllers.GetBook)

	books.Use(middleware.JWTAuthMiddleware())
	{
		books.POST("/", controllers.CreateBook)
		books.PUT("/:id", controllers.UpdateBook)
		books.DELETE("/:id", controllers.DeleteBook)
	}
}
