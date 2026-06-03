package feed

import (
    "pengaduan_be2/internal/middleware"
    "github.com/gin-gonic/gin"
)

func RegisterFeedRoutes(r *gin.RouterGroup, handler *FeedHandler) {
    feed := r.Group("/feed")
    feed.Use(middleware.AuthMiddleware())
    {
        feed.GET("", handler.GetFeed)
        feed.POST("/like", handler.LikePost)
        feed.POST("/save", handler.SavePost)
        feed.POST("/comment", handler.AddComment)
        feed.GET("/comments/:postId", handler.GetComments)
    }
}