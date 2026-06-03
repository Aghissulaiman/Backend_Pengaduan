package province

import (
    "pengaduan_be2/internal/middleware"
    "github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, handler *ProvinceHandler) {
    // Public routes (tanpa auth)
    public := r.Group("/")
    {
        public.GET("/provinces", handler.GetAllProvinces)
        public.GET("/provinces/:province_api_id/regencies", handler.GetRegencies)
        public.GET("/regencies/:regency_id/districts", handler.GetDistricts)
        public.GET("/districts/:district_id/villages", handler.GetVillages)
    }

    // Admin only routes
    admin := r.Group("/admin")
    admin.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))
    {
        admin.POST("/provinces/sync", handler.SyncProvinces)
        admin.POST("/regencies/sync", handler.SyncRegencies)
    }
}