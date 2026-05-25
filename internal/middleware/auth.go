package middleware

import (
	"fmt"
	"net/http"

	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		fmt.Println(authHeader)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, Response{
				Success: false,
				Message: "Token tidak ditemukan",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, Response{
				Success: false,
				Message: "Format token salah",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		secret := []byte(os.Getenv("JWT_SECRET"))

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return secret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, Response{
				Success: false,
				Message: "Token tidak valid",
			})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Tambahkan pengecekan type assertion yang aman
			if userID, ok := claims["user_id"].(float64); ok {
				c.Set("user_id", int(userID))
			}
			if username, ok := claims["username"].(string); ok {
				c.Set("username", username)
			}
			if role, ok := claims["role"].(string); ok {
				c.Set("role", role)
			}
			if provinceID, ok := claims["province_id"].(float64); ok {
				c.Set("province_id", int(provinceID))
			}
		}

		c.Next()
	}
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, Response{
				Success: false,
				Message: "Akses ditolak",
			})
			c.Abort()
			return
		}

		roleStr := role.(string)
		for _, allowed := range allowedRoles {
			if roleStr == allowed {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, Response{
			Success: false,
			Message: "Anda tidak memiliki akses ke resource ini",
		})
		c.Abort()
	}
}
