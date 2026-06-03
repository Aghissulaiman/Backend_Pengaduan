package feed

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "pengaduan_be2/internal/dto"
)

type FeedHandler struct {
    service *FeedService
}

func NewFeedHandler() *FeedHandler {
    return &FeedHandler{service: NewFeedService()}
}

// GetFeed - GET /api/feed
func (h *FeedHandler) GetFeed(c *gin.Context) {
    userID, _ := c.Get("user_id")

    var req FeedRequest
    req.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
    req.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))
    req.Type = c.DefaultQuery("type", "for-you")

    posts, total, err := h.service.GetFeedPosts(userID.(int), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: "Gagal mengambil data feed",
        })
        return
    }

    totalPages := (total + req.Limit - 1) / req.Limit

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Data: FeedResponse{
            Posts:      posts,
            Total:      total,
            Page:       req.Page,
            Limit:      req.Limit,
            TotalPages: totalPages,
        },
    })
}

// LikePost - POST /api/feed/like
func (h *FeedHandler) LikePost(c *gin.Context) {
    userID, _ := c.Get("user_id")

    var req LikeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    if err := h.service.LikePost(userID.(int), req.PostID); err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: "Gagal memproses like",
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Message: "OK",
    })
}

// SavePost - POST /api/feed/save
func (h *FeedHandler) SavePost(c *gin.Context) {
    userID, _ := c.Get("user_id")

    var req SaveRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    if err := h.service.SavePost(userID.(int), req.PostID); err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: "Gagal memproses save",
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Message: "OK",
    })
}

// AddComment - POST /api/feed/comment
func (h *FeedHandler) AddComment(c *gin.Context) {
    userID, _ := c.Get("user_id")

    var req CommentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    if err := h.service.AddComment(userID.(int), req.PostID, req.Text); err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: "Gagal menambah komentar",
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Message: "Komentar ditambahkan",
    })
}

// GetComments - GET /api/feed/comments/:postId
func (h *FeedHandler) GetComments(c *gin.Context) {
    postID, err := strconv.Atoi(c.Param("postId"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: "ID tidak valid",
        })
        return
    }

    comments, err := h.service.GetComments(postID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: "Gagal mengambil komentar",
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Data:    comments,
    })
}