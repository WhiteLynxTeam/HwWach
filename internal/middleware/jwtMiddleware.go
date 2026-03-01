package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Константы для ключей контекста
const (
	ContextKeyUserUUID = "user_uuid"
)

func JWTMiddleware(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenUnverifiable
			}
			return secret, nil
		})
		if err != nil || !token.Valid {
			c.Error(fmt.Errorf("JWT parse error: %w", err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Error(fmt.Errorf("JWT claims type assertion failed"))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Malformed token claims"})
			return
		}

		sub, ok := claims["sub"].(string)
		if !ok {
			c.Error(fmt.Errorf("JWT subject claim missing or invalid: %v", claims["sub"]))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid subject claim"})
			return
		}

		userUUID, err := uuid.Parse(sub)
		if err != nil {
			c.Error(fmt.Errorf("UUID parse error: %w", err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid UUID format"})
			return
		}

		c.Set(ContextKeyUserUUID, userUUID.String())
		c.Next()
	}
}

// GetUserUUID извлекает UUID пользователя из контекста
// Возвращает пустой uuid.UUID и false если пользователь не авторизован
func GetUserUUID(c *gin.Context) (uuid.UUID, bool) {
	userUUID, exists := c.Get(ContextKeyUserUUID)
	if !exists {
		return uuid.Nil, false
	}

	u, ok := userUUID.(string)
	if !ok {
		return uuid.Nil, false
	}

	parsed, err := uuid.Parse(u)
	if err != nil {
		return uuid.Nil, false
	}

	return parsed, true
}

// RequireUserUUID извлекает UUID пользователя или прерывает запрос с 401
func RequireUserUUID(c *gin.Context) (uuid.UUID, bool) {
	userUUID, ok := GetUserUUID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return uuid.Nil, false
	}
	return userUUID, true
}
