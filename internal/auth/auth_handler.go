package auth

import (
    "net/http"
    "pengaduan_be2/internal/dto"
	"fmt"
    "github.com/gin-gonic/gin"
)

type AuthHandler struct {
    service *AuthService
}

func NewAuthHandler() *AuthHandler {
    return &AuthHandler{service: NewAuthService()}
}

func (h *AuthHandler) getUserID(c *gin.Context) (int, error) {
    userIDRaw, exists := c.Get("user_id")
    if !exists || userIDRaw == nil {
        return 0, fmt.Errorf("user not authenticated")
    }

    userID, ok := userIDRaw.(int)
    if !ok {
        return 0, fmt.Errorf("invalid user ID format")
    }

    return userID, nil
}

func (h *AuthHandler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    if err := h.service.Register(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
   }

    c.JSON(http.StatusCreated, dto.Response{
        Success: true,
        Message: "Registrasi berhasil, silahkan login",
    })
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    resp, err := h.service.Login(&req)
    if err != nil {
        c.JSON(http.StatusUnauthorized, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Data:    resp,
    })
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
    var req GoogleLoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    resp, err := h.service.GoogleLogin(&req)
    if err != nil {
        c.JSON(http.StatusUnauthorized, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Data:    resp,
    })
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
    userID, err := h.getUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    user, err := h.service.GetUserByID(userID)
    if err != nil {
        c.JSON(http.StatusNotFound, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Data:    user,
    })
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
    userID, err := h.getUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    var req UpdateProfileRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    if err := h.service.UpdateProfile(userID, &req); err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    // Ambil data user terbaru
    user, err := h.service.GetUserByID(userID)
    if err != nil {
        c.JSON(http.StatusOK, dto.Response{
            Success: true,
            Message: "Profil berhasil diperbarui",
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Message: "Profil berhasil diperbarui",
        Data:    user,
    })
}

// GetUserProfile - GET /api/users/:username
func (h *AuthHandler) GetUserProfile(c *gin.Context) {
    currentUserID, err := h.getUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    username := c.Param("username")
    profile, err := h.service.GetUserProfileByUsername(currentUserID, username)
    if err != nil {
        c.JSON(http.StatusNotFound, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Data:    profile,
    })
}

// GetUserPosts - GET /api/users/:username/posts
func (h *AuthHandler) GetUserPosts(c *gin.Context) {
    currentUserID, err := h.getUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    username := c.Param("username")
    posts, err := h.service.GetUserPostsByUsername(currentUserID, username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: "Gagal mengambil postingan",
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Data:    posts,
    })
}