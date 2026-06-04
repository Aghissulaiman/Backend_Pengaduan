package chat

import (
    "pengaduan_be2/internal/middleware"
    "github.com/gin-gonic/gin"
)

func RegisterChatRoutes(r *gin.RouterGroup, handler *ChatHandler) {
    chat := r.Group("/chat")
    chat.Use(middleware.AuthMiddleware())
    {
        chat.GET("/conversations", handler.GetConversations)
        chat.GET("/messages/:conversationId", handler.GetMessages)
        chat.POST("/messages", handler.SendMessage)
        chat.GET("/users", handler.GetChatUsers)
        chat.POST("/conversation", handler.GetOrCreateConversation)
        chat.GET("/contacts", handler.GetContacts)
        chat.POST("/follow", handler.FollowUser)
        chat.POST("/follow/accept", handler.AcceptFollow)
        chat.POST("/follow/reject", handler.RejectFollow)
        chat.GET("/follow/requests", handler.GetFollowRequests)
    }
}