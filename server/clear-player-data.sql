-- 清理所有玩家相关数据的脚本
-- 注意：这将删除所有玩家的游戏进度，请谨慎使用！

-- 禁用外键约束检查（PostgreSQL）
SET session_replication_role = replica;

-- 删除所有玩家相关数据
DELETE FROM "Artifacts";
DELETE FROM "Pills";
DELETE FROM "Herbs";
DELETE FROM "Items";
DELETE FROM "Users";

-- 重置自增ID序列
SELECT setval(pg_get_serial_sequence('"Users"', 'id'), coalesce(max(id), 0) + 1, false) FROM "Users";
SELECT setval(pg_get_serial_sequence('"Items"', 'id'), coalesce(max(id), 0) + 1, false) FROM "Items";
SELECT setval(pg_get_serial_sequence('"Herbs"', 'id'), coalesce(max(id), 0) + 1, false) FROM "Herbs";
SELECT setval(pg_get_serial_sequence('"Pills"', 'id'), coalesce(max(id), 0) + 1, false) FROM "Pills";
SELECT setval(pg_get_serial_sequence('"Artifacts"', 'id'), coalesce(max(id), 0) + 1, false) FROM "Artifacts";

-- 重新启用外键约束检查
SET session_replication_role = DEFAULT;

-- 显示清理结果
SELECT '所有玩家数据已清理' as result;