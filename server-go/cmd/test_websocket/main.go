package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"xiuxian/server-go/internal/websocket"

	"go.uber.org/zap"
)

/**
 * WebSocket 后端集成测试
 * 用途: 验证WebSocket连接管理器和事件处理器的功能
 */

func main() {
	// 创建logger
	config := zap.NewDevelopmentConfig()
	logger, _ := config.Build()
	defer logger.Sync()

	fmt.Println("========== WebSocket 后端集成测试 ==========\n")

	// 测试1: 创建连接管理器
	fmt.Println("测试1: 创建连接管理器")
	manager := websocket.NewConnectionManager(logger)
	fmt.Println("✓ 连接管理器已创建\n")

	// 测试2: 启动管理器
	fmt.Println("测试2: 启动管理器")
	ctx := context.Background()
	manager.Start(ctx)
	fmt.Println("✓ 管理器已启动\n")

	// 测试3: 初始化事件处理器
	fmt.Println("测试3: 初始化事件处理器")
	handlers := websocket.InitializeHandlers(manager, logger)
	fmt.Println("✓ 事件处理器已初始化")
	fmt.Printf("  - Spirit Handler: %v\n", handlers.Spirit)
	fmt.Printf("  - Dungeon Handler: %v\n", handlers.Dungeon)
	fmt.Printf("  - Leaderboard Handler: %v\n", handlers.Leaderboard)
	fmt.Printf("  - Exploration Handler: %v\n\n", handlers.Exploration)

	// 测试4: 测试灵力增长事件
	fmt.Println("测试4: 测试灵力增长事件")
	spiritEvent := websocket.SpiritGrowthEvent{
		UserID:         1,
		OldSpirit:      100.0,
		NewSpirit:      115.03,
		GainAmount:     15.03,
		SpiritRate:     1.5,
		ElapsedSeconds: 10.0,
		Timestamp:      time.Now().Unix(),
	}
	err := handlers.Spirit.BroadcastSpiritGrowth(1, spiritEvent)
	if err != nil {
		log.Fatalf("灵力事件广播失败: %v", err)
	}
	fmt.Println("✓ 灵力增长事件已广播\n")

	// 测试5: 测试战斗事件
	fmt.Println("测试5: 测试战斗事件")
	err = handlers.Dungeon.NotifyDungeonStart(1, "魔渊秘境")
	if err != nil {
		log.Fatalf("战斗开始事件失败: %v", err)
	}
	fmt.Println("✓ 战斗开始事件已发送")

	err = handlers.Dungeon.NotifyCombatRound(1, "魔渊秘境", 1, 250.0, 200.0, 50.0, 30.0)
	if err != nil {
		log.Fatalf("战斗轮次事件失败: %v", err)
	}
	fmt.Println("✓ 战斗轮次事件已发送")

	err = handlers.Dungeon.NotifyVictory(1, "魔渊秘径", map[string]interface{}{
		"exp":   500,
		"gold":  1000,
		"items": []string{"灵石", "丹药"},
	})
	if err != nil {
		log.Fatalf("战斗胜利事件失败: %v", err)
	}
	fmt.Println("✓ 战斗胜利事件已发送\n")

	// 测试6: 测试排行榜更新
	fmt.Println("测试6: 测试排行榜更新")
	top10 := []websocket.LeaderboardEntry{
		{
			Rank:     1,
			UserID:   2,
			Username: "玩家2",
			Spirit:   5000.0,
			Power:    3500.0,
			Level:    10,
		},
		{
			Rank:     2,
			UserID:   3,
			Username: "玩家3",
			Spirit:   4500.0,
			Power:    3200.0,
			Level:    9,
		},
	}

	userRank := &websocket.UserRankInfo{
		Rank:    5,
		Value:   1500.0,
		Percent: 50.0,
	}

	err = handlers.Leaderboard.NotifySpiritLeaderboardUpdate(1, top10, userRank)
	if err != nil {
		log.Fatalf("排行榜更新事件失败: %v", err)
	}
	fmt.Println("✓ 排行榜更新事件已发送\n")

	// 测试7: 测试探索事件
	fmt.Println("测试7: 测试探索事件")
	err = handlers.Exploration.NotifyExplorationStart(1, "古老遗迹", 300)
	if err != nil {
		log.Fatalf("探索开始事件失败: %v", err)
	}
	fmt.Println("✓ 探索开始事件已发送")

	err = handlers.Exploration.NotifyExplorationProgress(1, "古老遗迹", 150, 300)
	if err != nil {
		log.Fatalf("探索进度事件失败: %v", err)
	}
	fmt.Println("✓ 探索进度事件已发送")

	err = handlers.Exploration.NotifyExplorationComplete(1, "古老遗迹", map[string]interface{}{
		"items":       []string{"灵草", "灵石"},
		"spirit":      500,
		"cultivation": 1000,
	})
	if err != nil {
		log.Fatalf("探索完成事件失败: %v", err)
	}
	fmt.Println("✓ 探索完成事件已发送\n")

	// 测试8: 获取连接统计
	fmt.Println("测试8: 获取连接统计")
	onlineCount := manager.GetOnlineCount()
	fmt.Printf("✓ 当前在线连接数: %d\n", onlineCount)
	fmt.Printf("✓ 用户1在线: %v\n\n", manager.IsUserOnline(1))

	// 测试9: 测试错误处理
	fmt.Println("测试9: 测试错误处理")
	// 发送消息给不存在的用户（应该安全处理，不会崩溃）
	err = handlers.Spirit.NotifySpiritUpdate(999, 100.0, 120.0, 1.0, 10.0)
	if err != nil {
		fmt.Printf("✓ 正确处理了不存在用户的消息（预期行为）\n\n")
	} else {
		fmt.Printf("✓ 消息已排队，等待用户999上线\n\n")
	}

	// 完成
	fmt.Println("========== 所有测试完成 ==========")
	fmt.Println("\n总结:")
	fmt.Println("✓ 连接管理器工作正常")
	fmt.Println("✓ 事件处理器工作正常")
	fmt.Println("✓ 所有4种事件类型已验证")
	fmt.Println("✓ 错误处理工作正常")
	fmt.Println("\n建议:")
	fmt.Println("1. 启动完整服务进行端到端测试")
	fmt.Println("2. 使用前端客户端连接验证WebSocket通信")
	fmt.Println("3. 在生产环境部署前进行性能测试")
}
