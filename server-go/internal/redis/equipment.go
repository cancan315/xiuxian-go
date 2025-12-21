package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// 装备强化和洗练的 Redis 操作工具

// 装备资源缓存键前缀
const (
	// 用户强化石数量缓存
	EquipmentResourceKeyFormat = "user:%d:equipment:resources"
	// 用户装备缓存
	EquipmentCacheKeyFormat = "user:%d:equipment:%s"
	// 用户装备列表缓存
	EquipmentListCacheKeyFormat = "user:%d:equipment:list"
	// 装备强化锁（防并发）
	EquipmentEnhanceLockKeyFormat = "user:%d:equipment:%s:enhance:lock"
	// 装备洗练锁（防并发）
	EquipmentReforgeLockKeyFormat = "user:%d:equipment:%s:reforge:lock"
	// ✅ 新增：装备操作最后修改时间戳 - 用于判断是否需要同步到DB
	EquipmentLastOperationTimeKeyFormat = "user:%d:equipment:last_operation_time"
	// 装备操作缓存过期时间（必须 > 心跳超时时间15s，确保离线同步数据不丢失）
	EquipmentCacheTTL = 20 * time.Second
	// 装备列表缓存过期时间
	EquipmentListCacheTTL = 5 * time.Second
	// 操作锁超时时间（必须 > 心跳超时时间15s）
	OperationLockTTL = 20 * time.Second
)

// EquipmentResources 装备资源（强化石、洗练石数量）
type EquipmentResources struct {
	ReinforceStones  int64 `json:"reinforce_stones"`
	RefinementStones int64 `json:"refinement_stones"`
	UpdatedAt        int64 `json:"updated_at"`
}

// GetEquipmentResources 从 Redis 获取用户的装备资源缓存
func GetEquipmentResources(ctx context.Context, userID uint) (*EquipmentResources, error) {
	key := fmt.Sprintf(EquipmentResourceKeyFormat, userID)
	val, err := Client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var resources EquipmentResources
	if err := json.Unmarshal([]byte(val), &resources); err != nil {
		return nil, err
	}

	return &resources, nil
}

// SetEquipmentResources 将装备资源缓存到 Redis
// reinforceStones: 强化石数量
// refinementStones: 洗练石数量
func SetEquipmentResources(ctx context.Context, userID uint, reinforceStones, refinementStones int64) error {
	key := fmt.Sprintf(EquipmentResourceKeyFormat, userID)
	resources := EquipmentResources{
		ReinforceStones:  reinforceStones,
		RefinementStones: refinementStones,
		UpdatedAt:        time.Now().Unix(),
	}

	data, err := json.Marshal(resources)
	if err != nil {
		return err
	}

	if err := Client.Set(ctx, key, string(data), EquipmentCacheTTL).Err(); err != nil {
		return err
	}

	// ✅ 新增：记录操作时间戳
	UpdateEquipmentLastOperationTime(ctx, userID)

	return nil
}

// DecrementReinforceStones 原子性地减少强化石数量（直接在 Redis 操作）
// 返回操作后的值
func DecrementReinforceStones(ctx context.Context, userID uint, amount int64) (int64, error) {
	key := fmt.Sprintf(EquipmentResourceKeyFormat, userID)

	// 获取当前值
	resources, err := GetEquipmentResources(ctx, userID)
	if err != nil {
		// 如果缓存不存在，返回错误
		return 0, err
	}

	// 减少强化石
	newValue := resources.ReinforceStones - amount
	resources.ReinforceStones = newValue
	resources.UpdatedAt = time.Now().Unix()

	// 更新回 Redis
	data, _ := json.Marshal(resources)
	err = Client.Set(ctx, key, string(data), EquipmentCacheTTL).Err()

	if err == nil {
		// ✅ 新增：记录操作时间戳
		UpdateEquipmentLastOperationTime(ctx, userID)
	}

	return newValue, err
}

// DecrementRefinementStones 原子性地减少洗练石数量
// 返回操作后的值
func DecrementRefinementStones(ctx context.Context, userID uint, amount int64) (int64, error) {
	key := fmt.Sprintf(EquipmentResourceKeyFormat, userID)

	// 获取当前值
	resources, err := GetEquipmentResources(ctx, userID)
	if err != nil {
		// 如果缓存不存在，返回错误
		return 0, err
	}

	// 减少洗练石
	newValue := resources.RefinementStones - amount
	resources.RefinementStones = newValue
	resources.UpdatedAt = time.Now().Unix()

	// 更新回 Redis
	data, _ := json.Marshal(resources)
	err = Client.Set(ctx, key, string(data), EquipmentCacheTTL).Err()

	if err == nil {
		// ✅ 新增：记录操作时间戳
		UpdateEquipmentLastOperationTime(ctx, userID)
	}

	return newValue, err
}

// InvalidateEquipmentListCache 清除用户的装备列表缓存
func InvalidateEquipmentListCache(ctx context.Context, userID uint) error {
	key := fmt.Sprintf(EquipmentListCacheKeyFormat, userID)
	return Client.Del(ctx, key).Err()
}

// InvalidateEquipmentCache 清除特定装备的缓存
func InvalidateEquipmentCache(ctx context.Context, userID uint, equipmentID string) error {
	key := fmt.Sprintf(EquipmentCacheKeyFormat, userID, equipmentID)
	return Client.Del(ctx, key).Err()
}

// TryEnhanceLock 尝试获取装备强化锁
// 成功返回 true，说明已获得锁；失败返回 false，说明已有其他强化在进行
func TryEnhanceLock(ctx context.Context, userID uint, equipmentID string) (bool, error) {
	key := fmt.Sprintf(EquipmentEnhanceLockKeyFormat, userID, equipmentID)
	// 使用 NX 选项确保只有在键不存在时才设置
	result, err := Client.SetNX(ctx, key, fmt.Sprintf("%d", time.Now().Unix()), OperationLockTTL).Result()
	return result, err
}

// ReleaseEnhanceLock 释放装备强化锁
func ReleaseEnhanceLock(ctx context.Context, userID uint, equipmentID string) error {
	key := fmt.Sprintf(EquipmentEnhanceLockKeyFormat, userID, equipmentID)
	return Client.Del(ctx, key).Err()
}

// TryReforgeLock 尝试获取装备洗练锁
func TryReforgeLock(ctx context.Context, userID uint, equipmentID string) (bool, error) {
	key := fmt.Sprintf(EquipmentReforgeLockKeyFormat, userID, equipmentID)
	result, err := Client.SetNX(ctx, key, fmt.Sprintf("%d", time.Now().Unix()), OperationLockTTL).Result()
	return result, err
}

// ReleaseReforgeLock 释放装备洗练锁
func ReleaseReforgeLock(ctx context.Context, userID uint, equipmentID string) error {
	key := fmt.Sprintf(EquipmentReforgeLockKeyFormat, userID, equipmentID)
	return Client.Del(ctx, key).Err()
}

// CacheEquipment 缓存装备数据到 Redis
func CacheEquipment(ctx context.Context, userID uint, equipmentID string, equipmentData interface{}) error {
	key := fmt.Sprintf(EquipmentCacheKeyFormat, userID, equipmentID)
	data, err := json.Marshal(equipmentData)
	if err != nil {
		return err
	}
	return Client.Set(ctx, key, string(data), EquipmentCacheTTL).Err()
}

// GetCachedEquipment 从 Redis 获取缓存的装备数据
func GetCachedEquipment(ctx context.Context, userID uint, equipmentID string) (map[string]interface{}, error) {
	key := fmt.Sprintf(EquipmentCacheKeyFormat, userID, equipmentID)
	val, err := Client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var equipmentData map[string]interface{}
	if err := json.Unmarshal([]byte(val), &equipmentData); err != nil {
		return nil, err
	}

	return equipmentData, nil
}

// SyncEquipmentResourcesToRedis 从数据库同步装备资源到 Redis
// 这个函数应该在用户登录或资源改变时调用
func SyncEquipmentResourcesToRedis(ctx context.Context, userID uint, reinforceStones, refinementStones int64) error {
	return SetEquipmentResources(ctx, userID, reinforceStones, refinementStones)
}

// ============================================
// ✅ 新增：操作时间戳管理
// ============================================

// UpdateEquipmentLastOperationTime 更新装备操作的最后修改时间戳
func UpdateEquipmentLastOperationTime(ctx context.Context, userID uint) error {
	key := fmt.Sprintf(EquipmentLastOperationTimeKeyFormat, userID)
	now := time.Now().Unix()
	return Client.Set(ctx, key, now, EquipmentCacheTTL).Err()
}

// GetEquipmentLastOperationTime 获取装备操作的最后修改时间戳
func GetEquipmentLastOperationTime(ctx context.Context, userID uint) (int64, error) {
	key := fmt.Sprintf(EquipmentLastOperationTimeKeyFormat, userID)
	val, err := Client.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	var timestamp int64
	fmt.Sscanf(val, "%d", &timestamp)
	return timestamp, nil
}

// ShouldSyncEquipmentToDb 判断装备资源是否需要同步到数据库
// 如果距离上次操作已超过5秒，返回 true
func ShouldSyncEquipmentToDb(ctx context.Context, userID uint, timeoutSeconds int64) bool {
	lastOpTime, err := GetEquipmentLastOperationTime(ctx, userID)
	if err != nil {
		// 如果获取时间戳失败，说明还没有操作过，不需要同步
		return false
	}

	elapsed := time.Now().Unix() - lastOpTime
	return elapsed >= timeoutSeconds
}
