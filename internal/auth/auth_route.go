package auth

import (
    "github.com/gin-gonic/gin"
    "pengaduan_be2/internal/middleware"
)

func RegisterRoutes(r *gin.RouterGroup, handler *AuthHandler) {
    auth := r.Group("/auth")
    {
        auth.POST("/register", handler.Register)
        auth.POST("/login", handler.Login)
        auth.POST("/google", handler.GoogleLogin)
    }

    user := r.Group("/users")
    user.Use(middleware.AuthMiddleware())
    {
        user.GET("/profile", handler.GetProfile)
        user.PUT("/profile", handler.UpdateProfile)
        user.GET("/:username", handler.GetUserProfile)        // 🔥 TAMBAHKAN
        user.GET("/:username/posts", handler.GetUserPosts)    // 🔥 TAMBAHKAN
    }
}