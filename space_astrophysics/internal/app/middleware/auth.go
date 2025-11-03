package middleware

import (
	"net/http"
	"strings"

	"space_astrophysics/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware проверяет JWT и допустимые роли
func AuthMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := utils.ParseJWT(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Сохраняем данные в контекст Gin
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)

		// Проверяем роль
		roleAllowed := false
		for _, r := range allowedRoles {
			if r == claims.Role {
				roleAllowed = true
				break
			}
		}

		if !roleAllowed && len(allowedRoles) > 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied for role: " + claims.Role})
			return
		}

		c.Next()
	}
}
