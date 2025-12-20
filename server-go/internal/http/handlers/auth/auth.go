package auth

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"

	"xiuxian/server-go/internal/db"
	playerHandler "xiuxian/server-go/internal/http/handlers/player"
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

	// 获取zap logger
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	// 记录注册请求日志
	zapLogger.Info("[注册] 收到注册请求",
		zap.String("username", req.Username))

	// 检查用户是否存在
	var existing models.User
	if err := db.DB.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		zapLogger.Warn("[注册] 用户名已存在",
			zap.String("username", req.Username))
		c.JSON(http.StatusBadRequest, gin.H{"message": "用户名已存在"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		zapLogger.Error("[注册] 密码加密失败",
			zap.String("username", req.Username),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	user := models.User{
		Username:          req.Username,
		Password:          string(hashed),
		PlayerName:        "无名修士",
		Level:             1,
		Realm:             "练气期一层",
		Cultivation:       0,
		MaxCultivation:    100,
		Spirit:            0,
		SpiritStones:      20000,
		ReinforceStones:   0,
		RefinementStones:  0,
		PetEssence:        0,
		BaseAttributes:    datatypes.JSON([]byte("{\"attack\":10,\"health\":100,\"defense\":5,\"speed\":10}")),
		CombatAttributes:  datatypes.JSON([]byte("{\"critRate\":0,\"comboRate\":0,\"counterRate\":0,\"stunRate\":0,\"dodgeRate\":0,\"vampireRate\":0}")),
		CombatResistance:  datatypes.JSON([]byte("{\"critResist\":0,\"comboResist\":0,\"counterResist\":0,\"stunResist\":0,\"dodgeResist\":0,\"vampireResist\":0}")),
		SpecialAttributes: datatypes.JSON([]byte("{\"healBoost\":0,\"critDamageBoost\":0,\"critDamageReduce\":0,\"finalDamageBoost\":0,\"finalDamageReduce\":0,\"combatBoost\":0,\"resistanceBoost\":0}")),
	}
	if err := db.DB.Create(&user).Error; err != nil {
		zapLogger.Error("[注册] 创建用户失败",
			zap.String("username", req.Username),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	zapLogger.Info("[注册] 用户创建成功,玩家数据初始化完成",
		zap.Uint("userID", user.ID),
		zap.String("username", user.Username),
		zap.Int("spirit_stones", user.SpiritStones))

	token, err := generateToken(user.ID)
	if err != nil {
		zapLogger.Error("[注册] 生成令牌失败",
			zap.Uint("userID", user.ID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	zapLogger.Info("[注册] 用户注册完成",
		zap.Uint("userID", user.ID),
		zap.String("username", user.Username))
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

	// 获取zap logger
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	// 记录登录请求日志
	zapLogger.Info("[登录] 收到登录请求",
		zap.String("username", req.Username))

	var user models.User
	if err := db.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		zapLogger.Warn("[登录] 用户不存在",
			zap.String("username", req.Username))
		c.JSON(http.StatusBadRequest, gin.H{"message": "用户名或密码错误"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		zapLogger.Warn("[登录] 密码错误",
			zap.String("username", req.Username))
		c.JSON(http.StatusBadRequest, gin.H{"message": "用户名或密码错误"})
		return
	}

	token, err := generateToken(user.ID)
	if err != nil {
		zapLogger.Error("[登录] 生成令牌失败",
			zap.Uint("userID", user.ID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// ✅ 新增：初始化装备资源缓存
	if err := playerHandler.InitEquipmentResourcesCache(c, user.ID); err != nil {
		zapLogger.Warn("[登录] 初始化装备资源缓存失败",
			zap.Uint("userID", user.ID),
			zap.Error(err))
		// 不中断流程，只记录日志
	}

	zapLogger.Info("[登录] 用户登录成功",
		zap.Uint("userID", user.ID),
		zap.String("username", user.Username))
	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"token":    token,
	})
}

// Logout 对应登出
func Logout(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	// 获取zap logger
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	// ✅ 新增：同步装备资源缓存到数据库
	if err := playerHandler.SyncEquipmentResourcesToDB(c, userID); err != nil {
		zapLogger.Warn("[登出] 同步装备资源到数据库失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		// 不中断流程，只记录日志
	}

	zapLogger.Info("[登出] 用户登出成功",
		zap.Uint("userID", userID))
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已登出"})
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

	// 获取zap logger
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	// 记录获取用户信息请求日志
	zapLogger.Info("[获取用户信息] 收到请求",
		zap.Uint("userID", userID))

	var user models.User
	if err := db.DB.Select("id", "username").First(&user, userID).Error; err != nil {
		zapLogger.Warn("[获取用户信息] 用户不存在",
			zap.Uint("userID", userID))
		c.JSON(http.StatusNotFound, gin.H{"message": "用户不存在"})
		return
	}

	zapLogger.Info("[获取用户信息] 成功",
		zap.Uint("userID", user.ID),
		zap.String("username", user.Username))
	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
	})
}
