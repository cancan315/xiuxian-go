package auth

import (
	"encoding/json"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"

	"xiuxian/server-go/internal/db"
	playerHandler "xiuxian/server-go/internal/http/handlers/player"
	"xiuxian/server-go/internal/models"
	"xiuxian/server-go/internal/redis"
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
		NameChangeCount:   0,
		Level:             1,
		Realm:             "练气期一层",
		Cultivation:       0,
		MaxCultivation:    100,
		Spirit:            0,
		SpiritStones:      0,
		ReinforceStones:   0,
		RefinementStones:  0,
		PetEssence:        0,
		BaseAttributes:    datatypes.JSON([]byte("{\"attack\":10,\"health\":100,\"defense\":5,\"speed\":10,\"spiritRate\":1.0,\"cultivationRate\":1.0}")),
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

	// ✅ 新增：注册后将玩家ID写入Redis自动增长列表
	playerIDStr := strconv.FormatUint(uint64(user.ID), 10)
	if err := redis.Client.SAdd(redis.Ctx, "spirit:auto:grow:players", playerIDStr).Err(); err != nil {
		zapLogger.Warn("[注册] 添加玩家到灵力自动增长列表失败",
			zap.Uint("userID", user.ID),
			zap.Error(err))
	} else {
		zapLogger.Info("[注册] 已将玩家添加到灵力自动增长列表",
			zap.Uint("userID", user.ID))
	}

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

	// ✅ 新增：登录时初始化属性并重新装备
	if err := playerHandler.InitializePlayerAttributesOnLogin(&user, zapLogger); err != nil {
		zapLogger.Error("[登录] 初始化玩家属性失败",
			zap.Uint("userID", user.ID),
			zap.Error(err))
		// 不中断登录流程，继续处理
	}

	token, err := generateToken(user.ID)
	if err != nil {
		zapLogger.Error("[登录] 生成令牌失败",
			zap.Uint("userID", user.ID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
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

// ✅ 新增：initializePlayerAttributesOnLogin 登录时初始化玩家属性
// 步骤：
// 1. 卸下所有装备
// 2. 召回所有灵宠
// 3. 计算基础属性（基于等级的公式）
// 4. 重新穿戴装备和灵宠
// 5. 保存到数据库
func initializePlayerAttributesOnLogin(user *models.User, zapLogger *zap.Logger) error {
	userID := user.ID

	// 步骤1：卸下所有装备
	if err := unequipAllEquipment(userID, zapLogger); err != nil {
		zapLogger.Error("[登录初始化] 卸下装备失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		// 不中断，继续执行
	}

	// 步骤2：召回所有灵宠
	if err := recallAllPets(userID, zapLogger); err != nil {
		zapLogger.Error("[登录初始化] 召回灵宠失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		// 不中断，继续执行
	}

	// 步骤3：计算基础属性
	baseAttrs := calculateBaseAttributes(user.Level)
	spiritRate := calculateSpiritRate(user.Level)
	cultivationRate := 1.0 // 默认修炼倍率

	// 步骤4：更新BaseAttributes
	baseAttrs["spiritRate"] = spiritRate
	baseAttrs["cultivationRate"] = cultivationRate

	attrJSON, err := json.Marshal(baseAttrs)
	if err != nil {
		return err
	}
	user.BaseAttributes = datatypes.JSON(attrJSON)

	// 步骤5：重新穿戴装备（获取玩家已有的装备）
	var equipments []models.Equipment
	if err := db.DB.Where("user_id = ? AND equipped = ?", userID, true).Find(&equipments).Error; err == nil {
		for _, equipment := range equipments {
			if err := reequipEquipmentAfterLogin(user, &equipment, zapLogger); err != nil {
				zapLogger.Warn("[登录初始化] 重新装备失败",
					zap.Uint("userID", userID),
					zap.String("equipmentID", equipment.ID),
					zap.Error(err))
			}
		}
	}

	// 步骤6：重新出战灵宠（获取玩家已有的灵宠）
	var pets []models.Pet
	if err := db.DB.Where("user_id = ? AND is_active = ?", userID, true).Find(&pets).Error; err == nil {
		for _, pet := range pets {
			if err := reactivatePetAfterLogin(user, &pet, zapLogger); err != nil {
				zapLogger.Warn("[登录初始化] 重新出战灵宠失败",
					zap.Uint("userID", userID),
					zap.String("petID", pet.PetID),
					zap.Error(err))
			}
		}
	}

	// 步骤7：保存更新后的用户属性到数据库
	if err := db.DB.Model(user).Updates(map[string]interface{}{
		"base_attributes":    user.BaseAttributes,
		"combat_attributes":  user.CombatAttributes,
		"combat_resistance":  user.CombatResistance,
		"special_attributes": user.SpecialAttributes,
	}).Error; err != nil {
		return err
	}

	zapLogger.Info("[登录初始化] 玩家属性初始化完成",
		zap.Uint("userID", userID),
		zap.Int("level", user.Level),
		zap.Any("baseAttributes", baseAttrs),
		zap.Float64("spiritRate", spiritRate))

	return nil
}

// ✅ 新增：calculateSpiritRate 计算灵力倍率
// 公式：spiritRate = 1.0 * (1.2)^(Level-1)
// 斗法消耗灵力公式 duelCost = 1440×1.2^(Level−1)
// 探索消耗灵力公式 exploreCost = 288×1.2^(Level−1)
func calculateSpiritRate(level int) float64 {
	if level < 1 {
		level = 1
	}
	spiritRate := 1.0 * math.Pow(1.2, float64(level-1))
	// 保留两位小数
	return math.Round(spiritRate*100) / 100
}

// ✅ 新增：calculateBaseAttributes 计算基础属性
// 基于等级的公式：
// speed = 10 * Level
// attack = 10 * Level
// health = 100 * Level
// defense = 5 * Level
func calculateBaseAttributes(level int) map[string]interface{} {
	return map[string]interface{}{
		"speed":   float64(10 * level),
		"attack":  float64(10 * level),
		"health":  float64(100 * level),
		"defense": float64(5 * level),
	}
}

// ✅ 新增：unequipAllEquipment 卸下玩家的所有装备
func unequipAllEquipment(userID uint, zapLogger *zap.Logger) error {
	var equipments []models.Equipment
	if err := db.DB.Where("user_id = ? AND equipped = ?", userID, true).Find(&equipments).Error; err != nil {
		return err
	}

	for _, equipment := range equipments {
		equipment.Equipped = false
		equipment.Slot = nil
		if err := db.DB.Save(&equipment).Error; err != nil {
			zapLogger.Warn("[卸装备] 卸下装备失败",
				zap.Uint("userID", userID),
				zap.String("equipmentID", equipment.ID),
				zap.Error(err))
		}
	}

	if len(equipments) > 0 {
		zapLogger.Info("[卸装备] 成功卸下所有装备",
			zap.Uint("userID", userID),
			zap.Int("count", len(equipments)))
	}

	return nil
}

// ✅ 新增：recallAllPets 召回玩家的所有灵宠
func recallAllPets(userID uint, zapLogger *zap.Logger) error {
	if err := db.DB.Model(&models.Pet{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Update("is_active", false).Error; err != nil {
		return err
	}

	zapLogger.Info("[召回灵宠] 成功召回所有灵宠",
		zap.Uint("userID", userID))

	return nil
}

// ✅ 新增：reequipEquipmentAfterLogin 登录后重新穿戴装备
// 需要重新计算装备属性加成
func reequipEquipmentAfterLogin(user *models.User, equipment *models.Equipment, zapLogger *zap.Logger) error {
	// 解析用户属性
	baseAttrs := jsonToFloatMap(user.BaseAttributes)
	combatAttrs := jsonToFloatMap(user.CombatAttributes)
	combatRes := jsonToFloatMap(user.CombatResistance)
	specialAttrs := jsonToFloatMap(user.SpecialAttributes)

	// 解析装备属性
	equipStats := jsonToFloatMap(equipment.Stats)

	// 使用属性管理器应用装备加成
	attrMgr := playerHandler.NewAttributeManager(baseAttrs, combatAttrs, combatRes, specialAttrs)
	attrMgr.ApplyEquipmentStats(equipStats)

	// 更新用户属性
	user.BaseAttributes = toJSONInterface(attrMgr.BaseAttrs)
	user.CombatAttributes = toJSONInterface(attrMgr.CombatAttrs)
	user.CombatResistance = toJSONInterface(attrMgr.CombatRes)
	user.SpecialAttributes = toJSONInterface(attrMgr.SpecialAttrs)

	zapLogger.Debug("[重新装备] 应用装备属性成功",
		zap.Uint("userID", user.ID),
		zap.String("equipmentID", equipment.ID))

	return nil
}

// ✅ 新增：reactivatePetAfterLogin 登录后重新出战灵宠
// 需要重新计算灵宠属性加成
func reactivatePetAfterLogin(user *models.User, pet *models.Pet, zapLogger *zap.Logger) error {
	// 解析用户属性
	baseAttrs := jsonToFloatMap(user.BaseAttributes)
	combatAttrs := jsonToFloatMap(user.CombatAttributes)
	combatRes := jsonToFloatMap(user.CombatResistance)
	specialAttrs := jsonToFloatMap(user.SpecialAttributes)

	// 解析灵宠战斗属性
	petCombat := jsonToFloatMap(pet.CombatAttributes)

	// 使用属性管理器应用灵宠加成
	attrMgr := playerHandler.NewAttributeManager(baseAttrs, combatAttrs, combatRes, specialAttrs)
	attrMgr.ApplyPetBonuses(pet, petCombat)

	// 更新用户属性
	user.BaseAttributes = toJSONInterface(attrMgr.BaseAttrs)
	user.CombatAttributes = toJSONInterface(attrMgr.CombatAttrs)
	user.CombatResistance = toJSONInterface(attrMgr.CombatRes)
	user.SpecialAttributes = toJSONInterface(attrMgr.SpecialAttrs)

	zapLogger.Debug("[重新出战灵宠] 应用灵宠属性成功",
		zap.Uint("userID", user.ID),
		zap.String("petID", pet.PetID))

	return nil
}

// ✅ 新增：jsonToFloatMap 将JSON转换为float64 map
func jsonToFloatMap(data datatypes.JSON) map[string]float64 {
	result := make(map[string]float64)
	if data != nil {
		var tempMap map[string]interface{}
		if err := json.Unmarshal(data, &tempMap); err == nil {
			for k, v := range tempMap {
				if f, ok := v.(float64); ok {
					result[k] = f
				}
			}
		}
	}
	return result
}

// ✅ 新增：toJSONInterface 将map[string]float64转换为JSON
func toJSONInterface(data map[string]float64) datatypes.JSON {
	// 转换为interface{} map
	interfaceMap := make(map[string]interface{})
	for k, v := range data {
		interfaceMap[k] = v
	}

	jsonBytes, err := json.Marshal(interfaceMap)
	if err != nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(jsonBytes)
}
