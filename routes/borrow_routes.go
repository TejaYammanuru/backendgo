package routes

import (
	"OnlineLibraryPortal/controllers"
	"OnlineLibraryPortal/middleware"

	"github.com/gin-gonic/gin"
)

func BorrowRoutes(router *gin.Engine) {
	borrow := router.Group("/borrow")
	borrow.Use(middleware.JWTAuthMiddleware())
	{
		// borrow.POST("/", controllers.BorrowBook)
		// borrow.POST("/return", controllers.ReturnBook)
		borrow.GET("/history", controllers.BorrowingHistory)
		borrow.GET("/librarians", controllers.GetAllLibrarians)
		borrow.GET("/members", controllers.GetAllMembers)
		borrow.GET("/dashboard", controllers.GetAdminDashboard)
		borrow.POST("/request", controllers.BorrowRequest)
		borrow.GET("/overdue", controllers.GetOverdueBooks)
		borrow.GET("/get-requests", controllers.GetAllBorrowRequests)
		borrow.POST("/approve", controllers.ApproveBorrowRequest)
		borrow.POST("/reject", controllers.RejectBorrowRequest)
		borrow.POST("/returnreq", controllers.ReturnRequest)
		borrow.POST("/returnack", controllers.AcknowledgeReturn)

	}
}
