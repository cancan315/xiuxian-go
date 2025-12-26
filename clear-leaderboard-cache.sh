#!/bin/bash

# 清除排行榜缓存脚本（通过 HTTP API）
# 用于清除服务器 Redis 中的排行榜缓存，使新的缓存数据结构生效
#
# 用法: ./clear-leaderboard-cache.sh [server_url]
# 示例: ./clear-leaderboard-cache.sh http://localhost:8080

# 默认服务器地址
SERVER_URL="${1:-http://localhost:8080}"
ENDPOINT="/api/admin/leaderboard/clear-cache"

echo "[清除脚本] 开始清除排行榜缓存..."
echo "[清除脚本] 目标服务器: $SERVER_URL"
echo ""

# 发送清除缓存请求
response=$(curl -s -X POST "$SERVER_URL$ENDPOINT" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json")

echo "[API 响应]"
echo "$response" | jq '.' 2>/dev/null || echo "$response"

echo ""
echo "[完成] 排行榜缓存清除请求已发送"
echo "[提示] 下次排行榜请求时将重新生成正确格式的缓存数据"
