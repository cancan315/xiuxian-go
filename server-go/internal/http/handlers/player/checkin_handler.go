package player

import (
	"encoding/json"
	"net/http"
	"time"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

// 签到奖励配置（第1-7天）
var checkInRewards = []int{1000, 2000, 3000, 4000, 5000, 6000, 10000}

// 从 BaseAttributes 获取签到数据
func getCheckInData(baseAttrs datatypes.JSON) (checkInDay int, lastCheckInDate string) {
	if len(baseAttrs) == 0 {
		return 0, ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal(baseAttrs, &m); err != nil {
		return 0, ""
	}
	if v, ok := m["checkInDay"].(float64); ok {
		checkInDay = int(v)
	}
	if v, ok := m["lastCheckInDate"].(string); ok {
		lastCheckInDate = v
	}
	return
}

// 更新 BaseAttributes 中的签到数据
func updateCheckInData(baseAttrs datatypes.JSON, checkInDay int, lastCheckInDate string) datatypes.JSON {
	var m map[string]interface{}
	if len(baseAttrs) == 0 {
		m = make(map[string]interface{})
	} else {
		if err := json.Unmarshal(baseAttrs, &m); err != nil {
			m = make(map[string]interface{})
		}
	}
	m["checkInDay"] = checkInDay
	m["lastCheckInDate"] = lastCheckInDate
	data, _ := json.Marshal(m)
	return datatypes.JSON(data)
}

// GetCheckInStatus 获取签到状态
func GetCheckInStatus(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "用户不存在"})
		return
	}

	// 从 BaseAttributes 获取签到数据
	checkInDay, lastCheckInDateStr := getCheckInData(user.BaseAttributes)

	// 检查今天是否已签到
	today := time.Now().Format("2006-01-02")
	hasCheckedInToday := today == lastCheckInDateStr

	// 检查是否断签（最后签到日期不是昨天或今天）
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	currentDay := checkInDay
	if !hasCheckedInToday && lastCheckInDateStr != yesterday && currentDay > 0 {
		// 断签，重置为0
		currentDay = 0
	}

	// 计算下次签到奖励
	nextDay := currentDay
	if hasCheckedInToday {
		nextDay = currentDay // 已签到，显示今天的奖励
	}
	if nextDay > 7 {
		nextDay = 1
	}
	if nextDay == 0 {
		nextDay = 1
	}

	c.JSON(http.StatusOK, gin.H{
		"success":           true,
		"checkInDay":        currentDay,                // 当前连续签到天数
		"hasCheckedInToday": hasCheckedInToday,         // 今天是否已签到
		"nextReward":        checkInRewards[nextDay-1], // 下次签到奖励
		"rewards":           checkInRewards,            // 所有奖励配置
	})
}

// DoCheckIn 执行签到
func DoCheckIn(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "用户不存在"})
		return
	}

	// 从 BaseAttributes 获取签到数据
	checkInDay, lastCheckInDateStr := getCheckInData(user.BaseAttributes)

	// 检查今天是否已签到
	today := time.Now().Format("2006-01-02")
	if today == lastCheckInDateStr {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "今日已签到"})
		return
	}

	// 检查是否断签
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	newDay := checkInDay + 1
	if lastCheckInDateStr != yesterday && checkInDay > 0 {
		// 断签，重置为第1天
		newDay = 1
	}
	// 超过7天循环
	if newDay > 7 {
		newDay = 1
	}

	// 计算奖励
	reward := checkInRewards[newDay-1]

	// 更新 BaseAttributes 中的签到数据
	newBaseAttrs := updateCheckInData(user.BaseAttributes, newDay, today)

	// 更新用户数据
	if err := db.DB.Model(&user).Updates(map[string]interface{}{
		"base_attributes": newBaseAttrs,
		"spirit_stones":   user.SpiritStones + reward,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "签到失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      "签到成功",
		"checkInDay":   newDay,
		"reward":       reward,
		"spiritStones": user.SpiritStones + reward,
	})
}
