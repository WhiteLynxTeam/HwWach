package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

// InternalKeyMiddleware внутренние запросы ts → go (5.1): общий ключ INTERNAL_API_KEY
// в заголовке X-Internal-Key, пользовательский токен не подходит
func InternalKeyMiddleware(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		got := c.GetHeader("X-Internal-Key")
		if key == "" || subtle.ConstantTimeCompare([]byte(got), []byte(key)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid internal key"})
			return
		}
		c.Next()
	}
}
