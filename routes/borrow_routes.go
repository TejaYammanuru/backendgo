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
		borrow.GET("/status", controllers.GetMyBorrowRequests)
		borrow.GET("/not-returned-books", controllers.GetBooksNotYetReturned)
		borrow.GET("/return-pending", controllers.GetBooksReturnRequestedNotAcknowledged)
		borrow.GET("/all-return-pending", controllers.GetReturnPendingRecords)
		borrow.GET("/lib-stats", controllers.GetLibrarianDashboardStats)
		borrow.GET("/member-notifications", controllers.GetMemberNotifications)
		borrow.GET("/member-overview", controllers.GetMemberOverview)

	}
}
