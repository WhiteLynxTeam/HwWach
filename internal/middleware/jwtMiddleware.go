package middleware

import (
	"HwWach/internal/authz"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const contextKeyPrincipal = "principal"

// Principal — authz.Principal: данные вошедшего из access-токена ts
type Principal = authz.Principal

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
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Malformed token claims"})
			return
		}

		// refresh-токен как access не подходит (5)
		if typ, _ := claims["typ"].(string); typ != "access" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token type"})
			return
		}

		p := &Principal{Perms: map[string]bool{}}
		var perr error
		if p.AccountID, perr = claimUUID(claims, "sub"); perr == nil {
			if p.OrgID, perr = claimUUID(claims, "org"); perr == nil {
				p.PersonID, perr = claimUUID(claims, "person")
			}
		}
		if perr != nil {
			c.Error(perr)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}
		p.Login, _ = claims["login"].(string)
		p.Role, _ = claims["role"].(string)
		p.IsMol, _ = claims["mol"].(bool)
		if perms, ok := claims["perms"].([]interface{}); ok {
			for _, v := range perms {
				if s, ok := v.(string); ok {
					p.Perms[s] = true
				}
			}
		}

		p.IP = c.ClientIP()
		c.Set(contextKeyPrincipal, p)
		c.Next()
	}
}

func claimUUID(claims jwt.MapClaims, key string) (uuid.UUID, error) {
	s, ok := claims[key].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("claim %s missing", key)
	}
	return uuid.Parse(s)
}

// CurrentPrincipal данные вошедшего или прерывает запрос с 401
func CurrentPrincipal(c *gin.Context) (*Principal, bool) {
	v, ok := c.Get(contextKeyPrincipal)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return nil, false
	}
	p, ok := v.(*Principal)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return nil, false
	}
	return p, true
}

// RequirePerm вошедший с любым из прав или 403
func RequirePerm(c *gin.Context, perms ...string) (*Principal, bool) {
	p, ok := CurrentPrincipal(c)
	if !ok {
		return nil, false
	}
	for _, perm := range perms {
		if p.Has(perm) {
			return p, true
		}
	}
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
	return nil, false
}
