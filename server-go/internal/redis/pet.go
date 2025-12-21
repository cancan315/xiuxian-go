package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// 灵宠升级升星的 Redis 操作工具

// 灵宠资源缓存键前缀
const (
	// 用户灵宠精华数量缓存
	PetResourceKeyFormat = "user:%d:pet:resources"
	// 用户灵宠缓存
	PetCacheKeyFormat = "user:%d:pet:%s"
	// 用户灵宠列表缓存
	PetListCacheKeyFormat = "user:%d:pet:list"
	// 灵宠升级锁（防并发）
	PetUpgradeLockKeyFormat = "user:%d:pet:%s:upgrade:lock"
	// 灵宠升星锁（防并发）
	PetEvolveLockKeyFormat = "user:%d:pet:%s:evolve:lock"
	// ✅ 新增：灵宠操作最后修改时间戳 - 用于判断是否需要同步到DB
	PetLastOperationTimeKeyFormat = "user:%d:pet:last_operation_time"
	// 灵宠操作缓存过期时间（必须 > 心跳超时时间15s，确保离线同步数据不丢失）
	PetCacheTTL = 20 * time.Second
	// 灵宠列表缓存过期时间
	PetListCacheTTL = 5 * time.Second
	// 灵宠操作锁超时时间（必须 > 心跳超时时间15s）
	PetOperationLockTTL = 20 * time.Second
)

// PetResources 灵宠资源（精华数量）
type PetResources struct {
	PetEssence int64 `json:"pet_essence"`
	UpdatedAt  int64 `json:"updated_at"`
}

// GetPetResources 从 Redis 获取用户的灵宠资源缓存
func GetPetResources(ctx context.Context, userID uint) (*PetResources, error) {
	key := fmt.Sprintf(PetResourceKeyFormat, userID)
	val, err := Client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var resources PetResources
	if err := json.Unmarshal([]byte(val), &resources); err != nil {
		return nil, err
	}

	return &resources, nil
}

// SetPetResources 将灵宠资源缓存到 Redis
func SetPetResources(ctx context.Context, userID uint, petEssence int64) error {
	key := fmt.Sprintf(PetResourceKeyFormat, userID)
	resources := PetResources{
		PetEssence: petEssence,
		UpdatedAt:  time.Now().Unix(),
	}

	data, err := json.Marshal(resources)
	if err != nil {
		return err
	}

	if err := Client.Set(ctx, key, string(data), PetCacheTTL).Err(); err != nil {
		return err
	}

	// ✅ 新增：记录操作时间戳
	UpdatePetLastOperationTime(ctx, userID)

	return nil
}

// DecrementPetEssence 原子性地减少灵宠精华数量
// 返回操作后的值
func DecrementPetEssence(ctx context.Context, userID uint, amount int64) (int64, error) {
	key := fmt.Sprintf(PetResourceKeyFormat, userID)

	// 获取当前值
	resources, err := GetPetResources(ctx, userID)
	if err != nil {
		// 如果缓存不存在，返回错误
		return 0, err
	}

	// 减少灵宠精华
	newValue := resources.PetEssence - amount
	resources.PetEssence = newValue
	resources.UpdatedAt = time.Now().Unix()

	// 更新回 Redis
	data, _ := json.Marshal(resources)
	err = Client.Set(ctx, key, string(data), PetCacheTTL).Err()

	if err == nil {
		// ✅ 新增：记录操作时间戳
		UpdatePetLastOperationTime(ctx, userID)
	}

	return newValue, err
}

// InvalidatePetListCache 清除用户的灵宠列表缓存
func InvalidatePetListCache(ctx context.Context, userID uint) error {
	key := fmt.Sprintf(PetListCacheKeyFormat, userID)
	return Client.Del(ctx, key).Err()
}

// InvalidatePetCache 清除特定灵宠的缓存
func InvalidatePetCache(ctx context.Context, userID uint, petID string) error {
	key := fmt.Sprintf(PetCacheKeyFormat, userID, petID)
	return Client.Del(ctx, key).Err()
}

// TryUpgradeLock 尝试获取灵宠升级锁
// 成功返回 true，说明已获得锁；失败返回 false，说明已有其他升级在进行
func TryUpgradeLock(ctx context.Context, userID uint, petID string) (bool, error) {
	key := fmt.Sprintf(PetUpgradeLockKeyFormat, userID, petID)
	// 使用 NX 选项确保只有在键不存在时才设置
	result, err := Client.SetNX(ctx, key, fmt.Sprintf("%d", time.Now().Unix()), PetOperationLockTTL).Result()
	return result, err
}

// ReleaseUpgradeLock 释放灵宠升级锁
func ReleaseUpgradeLock(ctx context.Context, userID uint, petID string) error {
	key := fmt.Sprintf(PetUpgradeLockKeyFormat, userID, petID)
	return Client.Del(ctx, key).Err()
}

// TryEvolveLock 尝试获取灵宠升星锁
func TryEvolveLock(ctx context.Context, userID uint, petID string) (bool, error) {
	key := fmt.Sprintf(PetEvolveLockKeyFormat, userID, petID)
	result, err := Client.SetNX(ctx, key, fmt.Sprintf("%d", time.Now().Unix()), PetOperationLockTTL).Result()
	return result, err
}

// ReleaseEvolveLock 释放灵宠升星锁
func ReleaseEvolveLock(ctx context.Context, userID uint, petID string) error {
	key := fmt.Sprintf(PetEvolveLockKeyFormat, userID, petID)
	return Client.Del(ctx, key).Err()
}

// CachePet 缓存灵宠数据到 Redis
func CachePet(ctx context.Context, userID uint, petID string, petData interface{}) error {
	key := fmt.Sprintf(PetCacheKeyFormat, userID, petID)
	data, err := json.Marshal(petData)
	if err != nil {
		return err
	}
	return Client.Set(ctx, key, string(data), PetCacheTTL).Err()
}

// GetCachedPet 从 Redis 获取缓存的灵宠数据
func GetCachedPet(ctx context.Context, userID uint, petID string) (map[string]interface{}, error) {
	key := fmt.Sprintf(PetCacheKeyFormat, userID, petID)
	val, err := Client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var petData map[string]interface{}
	if err := json.Unmarshal([]byte(val), &petData); err != nil {
		return nil, err
	}

	return petData, nil
}

// SyncPetResourcesToRedis 从数据库同步灵宠资源到 Redis
// 这个函数应该在用户登录或资源改变时调用
func SyncPetResourcesToRedis(ctx context.Context, userID uint, petEssence int64) error {
	return SetPetResources(ctx, userID, petEssence)
}

// ============================================
// ✅ 新增：操作时间戳管理
// ============================================

// UpdatePetLastOperationTime 更新灵宠操作的最后修改时间戳
func UpdatePetLastOperationTime(ctx context.Context, userID uint) error {
	key := fmt.Sprintf(PetLastOperationTimeKeyFormat, userID)
	now := time.Now().Unix()
	return Client.Set(ctx, key, now, PetCacheTTL).Err()
}

// GetPetLastOperationTime 获取灵宠操作的最后修改时间戳
func GetPetLastOperationTime(ctx context.Context, userID uint) (int64, error) {
	key := fmt.Sprintf(PetLastOperationTimeKeyFormat, userID)
	val, err := Client.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	var timestamp int64
	fmt.Sscanf(val, "%d", &timestamp)
	return timestamp, nil
}

// ShouldSyncPetToDb 判断灵宠资源是否需要同步到数据库
// 如果距离上次操作已超过5秒，返回 true
func ShouldSyncPetToDb(ctx context.Context, userID uint, timeoutSeconds int64) bool {
	lastOpTime, err := GetPetLastOperationTime(ctx, userID)
	if err != nil {
		// 如果获取时间戳失败，说明还没有操作过，不需要同步
		return false
	}

	elapsed := time.Now().Unix() - lastOpTime
	return elapsed >= timeoutSeconds
}
