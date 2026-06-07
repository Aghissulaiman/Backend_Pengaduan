package complaint

import (
	"github.com/gin-gonic/gin"
	"pengaduan_be2/internal/middleware"
)

func RegisterRoutes(r *gin.RouterGroup, handler *ComplaintHandler) {
	// Public routes (tanpa auth)
	public := r.Group("/complaints")
	{
		public.GET("/categories", handler.GetCategories)
		public.GET("/status/:tracking_code", handler.CheckStatus)
	}

	// User routes (semua role yang login)
	user := r.Group("/complaints")
	user.Use(middleware.AuthMiddleware())
	{
		user.POST("/submit", handler.SubmitComplaint)
		user.GET("/my", handler.GetMyComplaints)
		user.GET("/:id", handler.GetComplaintDetail)
		user.PATCH("/:id/status", handler.UpdateComplaintStatus)
	}

	// Routes untuk governor & admin & investigator (lihat semua pengaduan)
	all := r.Group("/complaints")
	all.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("governor", "admin", "investigator"))
	{
		all.GET("/all", handler.GetAllComplaints)
	}

	// Governor only routes
	governor := r.Group("/governor")
	governor.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("governor"))
	{
		// Dashboard & Stats
		governor.GET("/dashboard/stats", handler.GetGovernorStats)
		governor.GET("/investigators", handler.GetInvestigators)
		governor.GET("/investigations", handler.GetGovernorInvestigations)
		governor.GET("/reports", handler.GetGovernorReports) // 🔥 TAMBAHKAN ROUTE LAPORAN

		// Complaint actions
		governor.GET("/complaints", handler.GetGovernorComplaints)
		governor.POST("/complaints/:id/assign", handler.AssignInvestigator)
		governor.POST("/complaints/:id/process-report", handler.SubmitProcessReport)
		governor.POST("/complaints/:id/completion-report", handler.SubmitCompletionReport)
	}

	// Investigator only routes
	investigator := r.Group("/investigator")
	investigator.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("investigator"))
	{
		// Dashboard & Stats
		investigator.GET("/dashboard/stats", handler.GetInvestigatorStats)
		investigator.GET("/complaints", handler.GetInvestigatorComplaints)
		investigator.GET("/complaints/:id", handler.GetComplaintDetail)

		// Investigation actions
		investigator.POST("/complaints/:id/result", handler.SubmitInvestigationResultExtended)
	}

	// Admin only routes
	admin := r.Group("/admin")
	admin.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))
	{
		admin.GET("/dashboard", handler.GetDashboardStats)
		admin.GET("/dashboard/stats", handler.GetDashboardStats)
		admin.GET("/dashboard/charts", handler.GetDashboardCharts) // 🔥 TAMBAHKAN
		admin.GET("/complaints/count", handler.GetComplaintsCountByProvince)
		admin.GET("/complaints/by-province", handler.GetComplaintsByProvince)

		// User management
		admin.GET("/users", handler.GetAllUsers)
		admin.GET("/users/recent", handler.GetRecentUsers)
		admin.PUT("/users/:id", handler.UpdateUser)
		admin.PATCH("/users/:id/toggle-active", handler.ToggleUserActive)
		admin.GET("/provinces-with-users", handler.GetProvincesWithUsers)

		// Complaint management
		admin.GET("/complaints", handler.GetAllComplaintsForAdmin)
		admin.GET("/complaints/recent", handler.GetRecentComplaints)
		admin.GET("/complaints/:id", handler.GetComplaintDetailForAdmin)

		// Verification
		admin.GET("/process-reports", handler.GetProcessReports)
		admin.GET("/completion-reports", handler.GetCompletionReports)
		admin.POST("/process-reports/:id/verify", handler.VerifyProcessReport)
		admin.POST("/completion-reports/:id/verify", handler.VerifyCompletionReport)

		// Admin category management
		admin.GET("/categories", handler.GetCategoriesForAdmin)
		admin.POST("/categories", handler.CreateCategory)
		admin.PUT("/categories/:id", handler.UpdateCategory)
		admin.DELETE("/categories/:id", handler.DeleteCategory)
		admin.PATCH("/categories/:id/toggle-active", handler.ToggleCategoryActive)

		// Admin reports
		admin.GET("/reports/stats", handler.GetReportStats)
		admin.GET("/reports/export", handler.ExportReport)

		// Admin activity logs
		admin.GET("/activities", handler.GetActivityLogs)
		admin.GET("/activities/stats", handler.GetActivityStats)
	}
}
