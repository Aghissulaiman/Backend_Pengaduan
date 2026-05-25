package complaint

import (
    "pengaduan_be2/internal/middleware"
    "github.com/gin-gonic/gin"
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
    }

    // Routes untuk governor & admin & investigator (lihat semua pengaduan)
    all := r.Group("/complaints")
    all.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("governor", "admin", "investigator"))
    {
        all.GET("/all", handler.GetAllComplaints)
    }

    // Governor only routes
    governor := r.Group("/governor/complaints")
    governor.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("governor"))
    {
        governor.POST("/:id/assign", handler.AssignInvestigator)
        governor.POST("/:id/process-report", handler.SubmitProcessReport)
        governor.POST("/:id/completion-report", handler.SubmitCompletionReport)
    }

    // Investigator only routes
    investigator := r.Group("/investigator/complaints")
    investigator.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("investigator"))
    {
        investigator.POST("/:id/result", handler.SubmitInvestigationResult)
    }

    // Admin only routes
    admin := r.Group("/admin")
    admin.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))
    {
        admin.GET("/dashboard", handler.GetDashboardStats)
    }
}