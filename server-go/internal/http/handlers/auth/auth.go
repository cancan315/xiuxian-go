package auth

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
)

type jwtClaims struct {
	ID uint `json:"id"`
	jwt.RegisteredClaims
}

func generateToken(userID uint) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", os.ErrInvalid
	}

	claims := jwtClaims{
		ID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// Register 对应 Node 的 register
func Register(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误"})
		return
	}

	fmt.Printf("[注册] 收到注册请求，用户名: %s\n", req.Username)

	// 检查用户是否存在
	var existing models.User
	if err := db.DB.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		fmt.Printf("[注册] 用户名已存在: %s\n", req.Username)
		c.JSON(http.StatusBadRequest, gin.H{"message": "用户名已存在"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		fmt.Printf("[注册] 密码加密失败，用户名: %s, 错误: %v\n", req.Username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	user := models.User{
		Username:     req.Username,
		Password:     string(hashed),
		Level:        1,
		SpiritStones: 20000,
	}
	if err := db.DB.Create(&user).Error; err != nil {
		fmt.Printf("[注册] 创建用户失败，用户名: %s, 错误: %v\n", req.Username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	fmt.Printf("[注册] 用户创建成功，用户ID: %d, 用户名: %s\n", user.ID, user.Username)

	token, err := generateToken(user.ID)
	if err != nil {
		fmt.Printf("[注册] 生成令牌失败，用户ID: %d, 错误: %v\n", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	fmt.Printf("[注册] 用户注册完成，用户ID: %d, 用户名: %s\n", user.ID, user.Username)
	c.JSON(http.StatusCreated, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"token":    token,
	})
}

// Login 对应 Node 的 login
func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误"})
		return
	}

	fmt.Printf("[登录] 收到登录请求，用户名: %s\n", req.Username)

	var user models.User
	if err := db.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		fmt.Printf("[登录] 用户不存在，用户名: %s\n", req.Username)
		c.JSON(http.StatusBadRequest, gin.H{"message": "用户名或密码错误"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		fmt.Printf("[登录] 密码错误，用户名: %s\n", req.Username)
		c.JSON(http.StatusBadRequest, gin.H{"message": "用户名或密码错误"})
		return
	}

	token, err := generateToken(user.ID)
	if err != nil {
		fmt.Printf("[登录] 生成令牌失败，用户ID: %d, 错误: %v\n", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	fmt.Printf("[登录] 用户登录成功，用户ID: %d, 用户名: %s\n", user.ID, user.Username)
	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"token":    token,
	})
}

// GetUser 对应 Node 的 getUser
func GetUser(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	var user models.User
	if err := db.DB.Select("id", "username").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
	})
}
