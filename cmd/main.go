package main

import (
    "log"
    "os"
    "pengaduan_be2/internal/auth"
    "pengaduan_be2/internal/complaint"
    "pengaduan_be2/internal/middleware"
    "pengaduan_be2/internal/province"
    "pengaduan_be2/pkg/db"
    
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    
    // Swagger
    _ "pengaduan_be2/docs"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Pengaduan Provinsi API
// @version 1.0
// @description API untuk sistem pengaduan masyarakat tingkat provinsi
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
    // Load .env
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    // Init database
    if err := db.InitDB(); err != nil {
        log.Fatal("Database error:", err)
    }
    defer db.CloseDB()

    // Create uploads folder
    os.MkdirAll("uploads", 0755)

    // Init handlers
    authHandler := auth.NewAuthHandler()
    complaintHandler := complaint.NewComplaintHandler()
    provinceHandler := province.NewProvinceHandler()

    // Setup router
    r := gin.Default()
    r.Use(middleware.CORSMiddleware())
    r.Static("/uploads", "./uploads")

    // Swagger
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    // Health check
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // Register routes
    api := r.Group("/api")
    auth.RegisterRoutes(api, authHandler)
    complaint.RegisterRoutes(api, complaintHandler)
    province.RegisterRoutes(api, provinceHandler)

    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("🚀 Server running on port %s", port)
    log.Printf("📚 Swagger: http://localhost:%s/swagger/index.html", port)
    
    if err := r.Run(":" + port); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}