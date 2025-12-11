package gacha

import (
	"go.uber.org/zap"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
)

// GenerateEquipment 将生成的装备保存到数据库
func GenerateEquipment(userID uint, level int, logger *zap.Logger) (*models.Equipment, error) {
	eq := GenerateRandomEquipment(level)

	// 打印装备存入数据库前的属性
	logger.Info("[抽奖] 装备存入数据库前",
		zap.Uint("用户ID", userID),
		zap.String("装备名称", eq.Name),
		zap.String("品质", eq.Quality),
		zap.Any("属性", eq.Stats))

	model := models.Equipment{
		ID:              uuid.NewString(),
		UserID:          userID,
		EquipmentID:     eq.ID,
		Name:            eq.Name,
		Type:            eq.Type,
		Quality:         eq.Quality,
		EnhanceLevel:    eq.EnhanceLevel,
		RequiredRealm:   eq.RequiredRealm,
		Level:           eq.Level,
		EquipType:       &eq.EquipType,
		Stats:           ToJSON(eq.Stats),
		ExtraAttributes: ToJSON(eq.ExtraAttrs),
		Equipped:        false,
	}

	if err := db.DB.Create(&model).Error; err != nil {
		return nil, err
	}

	// 打印装备存入数据库后的属性
	logger.Info("[抽奖] 装备存入数据库后",
		zap.Uint("用户ID", userID),
		zap.String("装备ID", model.ID),
		zap.String("装备名称", model.Name),
		zap.String("品质", model.Quality),
		zap.Any("属性", model.Stats))

	return &model, nil
}

// GeneratePet 将生成的宠物保存到数据库
func GeneratePet(userID uint, level int, logger *zap.Logger) (*models.Pet, error) {
	p := GenerateRandomPet(level)

	// 打印灵宠存入数据库前的属性
	logger.Info("[抽奖] 灵宠存入数据库前",
		zap.Uint("用户ID", userID),
		zap.String("灵宠名称", p.Name),
		zap.String("稀有度", p.Rarity),
		zap.Any("属性", p.CombatAttrs))

	// 确保宠物ID唯一性
	petID := p.ID
	var existingPet models.Pet
	for {
		if err := db.DB.Where("user_id = ? AND pet_id = ?", userID, petID).First(&existingPet).Error; err != nil {
			break // ID未被占用，可以使用
		}
		petID = uuid.NewString() // ID已被占用，生成新的ID
	}

	petModel := models.Pet{
		ID:               uuid.NewString(),
		UserID:           userID,
		PetID:            petID,
		Name:             p.Name,
		Type:             p.Type,
		Rarity:           p.Rarity,
		Level:            p.Level,
		Star:             p.Star,
		Experience:       p.Exp,
		CombatAttributes: ToJSON(p.CombatAttrs),
		AttackBonus:      p.AttackBonus,
		DefenseBonus:     p.DefenseBonus,
		HealthBonus:      p.HealthBonus,
		IsActive:         false,
	}

	if err := db.DB.Create(&petModel).Error; err != nil {
		return nil, err
	}

	// 打印灵宠存入数据库后的属性
	logger.Info("[抽奖] 灵宠存入数据库后",
		zap.Uint("用户ID", userID),
		zap.String("灵宠ID", petModel.ID),
		zap.String("灵宠名称", petModel.Name),
		zap.String("稀有度", petModel.Rarity),
		zap.Any("属性", petModel.CombatAttributes))

	return &petModel, nil
}

// GetUser 获取用户信息
func GetUser(userID uint) (*models.User, error) {
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// DeductSpiritStones 扣除用户灵石
func DeductSpiritStones(userID uint, amount int) error {
	return db.DB.Model(&models.User{}).Where("id = ?", userID).
		Update("spirit_stones", gorm.Expr("spirit_stones - ?", amount)).Error
}

// AddReinforceStones 增加用户强化石
func AddReinforceStones(userID uint, amount int) error {
	return db.DB.Model(&models.User{}).Where("id = ?", userID).
		Update("reinforce_stones", gorm.Expr("reinforce_stones + ?", amount)).Error
}

// DeleteEquipment 删除指定装备
func DeleteEquipment(userID uint, equipmentID string) error {
	return db.DB.Where("user_id = ? AND equipment_id = ?", userID, equipmentID).
		Delete(&models.Equipment{}).Error
}

// DeletePet 删除指定宠物
func DeletePet(userID uint, petID string) error {
	return db.DB.Where("user_id = ? AND pet_id = ?", userID, petID).
		Delete(&models.Pet{}).Error
}