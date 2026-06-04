package chat

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "pengaduan_be2/internal/dto"
)

type ChatHandler struct {
    service *ChatService
}

func NewChatHandler() *ChatHandler {
    return &ChatHandler{service: NewChatService()}
}

// GetConversations - GET /api/chat/conversations
func (h *ChatHandler) GetConversations(c *gin.Context) {
    userID, _ := c.Get("user_id")

    var query GetConversationsQuery
    query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
    query.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))

    conversations, total, err := h.service.GetUserConversations(userID.(int), &query)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: "Gagal mengambil percakapan",
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Data: gin.H{
            "conversations": conversations,
            "total":         total,
            "page":          query.Page,
            "limit":         query.Limit,
        },
    })
}

// GetMessages - GET /api/chat/messages/:conversationId
func (h *ChatHandler) GetMessages(c *gin.Context) {
    userID, _ := c.Get("user_id")

    conversationID, err := strconv.Atoi(c.Param("conversationId"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: "ID tidak valid",
        })
        return
    }

    var query GetMessagesQuery
    query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
    query.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "50"))

    messages, total, err := h.service.GetMessages(conversationID, userID.(int), &query)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Data: gin.H{
            "messages": messages,
            "total":    total,
            "page":     query.Page,
            "limit":    query.Limit,
        },
    })
}

// SendMessage - POST /api/chat/messages
func (h *ChatHandler) SendMessage(c *gin.Context) {
    userID, _ := c.Get("user_id")

    var req SendMessageRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    if req.ReceiverID == userID.(int) {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: "Tidak bisa mengirim pesan ke diri sendiri",
        })
        return
    }

    message, err := h.service.SendMessage(userID.(int), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Message: "Pesan terkirim",
        Data:    message,
    })
}

// GetChatUsers - GET /api/chat/users
func (h *ChatHandler) GetChatUsers(c *gin.Context) {
    userID, _ := c.Get("user_id")
    search := c.Query("search")

    users, err := h.service.GetChatUsers(userID.(int), search)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: "Gagal mengambil daftar user",
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Data:    users,
    })
}

// GetOrCreateConversation - POST /api/chat/conversation
func (h *ChatHandler) GetOrCreateConversation(c *gin.Context) {
    userID, _ := c.Get("user_id")

    var req struct {
        UserID int `json:"user_id" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    conv, err := h.service.GetOrCreateConversation(userID.(int), req.UserID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Data:    conv,
    })
}

// GetContacts - GET /api/chat/contacts
func (h *ChatHandler) GetContacts(c *gin.Context) {
    userID, _ := c.Get("user_id")
    search := c.Query("search")

    contacts, err := h.service.GetContacts(userID.(int), search) // Hanya 2 nilai balik
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: "Gagal mengambil kontak: " + err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Data:    contacts,
    })
}

// FollowUser - POST /api/chat/follow
func (h *ChatHandler) FollowUser(c *gin.Context) {
    userID, _ := c.Get("user_id")

    var req struct {
        FollowingID int `json:"following_id" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    if err := h.service.FollowUser(userID.(int), req.FollowingID); err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Message: "Permintaan follow dikirim",
    })
}

// AcceptFollow - POST /api/chat/follow/accept
func (h *ChatHandler) AcceptFollow(c *gin.Context) {
    userID, _ := c.Get("user_id")

    var req struct {
        FollowerID int `json:"follower_id" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    if err := h.service.AcceptFollow(req.FollowerID, userID.(int)); err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Message: "Permintaan follow diterima",
    })
}

// RejectFollow - POST /api/chat/follow/reject
func (h *ChatHandler) RejectFollow(c *gin.Context) {
    userID, _ := c.Get("user_id")

    var req struct {
        FollowerID int `json:"follower_id" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    if err := h.service.RejectFollow(req.FollowerID, userID.(int)); err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Message: "Permintaan follow ditolak",
    })
}

// GetFollowRequests - GET /api/chat/follow/requests
func (h *ChatHandler) GetFollowRequests(c *gin.Context) {
    userID, _ := c.Get("user_id")

    requests, err := h.service.GetFollowRequests(userID.(int))
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: "Gagal mengambil permintaan follow",
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Data:    requests,
    })
}