package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type Claims struct {
	ID uint `json:"id"`
	jwt.RegisteredClaims
}

// Protect 与 Node 版本的 authMiddleware.protect 对齐：
// - 从 Authorization: Bearer <token> 读取
// - 校验 JWT_SECRET
// - 失败返回 401
func Protect() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		// ✅ 添加认证请求日志
		zap.L().Info("认证中间件处理请求",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("authHeader", authHeader))

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			// ✅ 记录认证失败原因
			zap.L().Warn("认证失败：缺少或无效的认证头",
				zap.String("path", c.Request.URL.Path),
				zap.String("authHeader", authHeader))
			c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权，没有令牌"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": "JWT_SECRET 未配置"})
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			// ✅ 记录JWT验证失败原因
			zap.L().Warn("认证失败：JWT验证失败",
				zap.String("path", c.Request.URL.Path),
				zap.String("tokenString", tokenString),
				zap.Error(err),
				zap.Bool("tokenValid", token != nil && token.Valid))
			c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
			c.Abort()
			return
		}

		// 保存用户 ID 到上下文，供后续 handler 使用
		c.Set("userID", claims.ID)
		c.Next()
	}
}
