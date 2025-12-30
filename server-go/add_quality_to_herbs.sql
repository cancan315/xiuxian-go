-- 创建迁移文件：add_quality_to_herbs.sql
-- 这个脚本应该在生产环境中执行以添加品质字段

ALTER TABLE "herbs" ADD COLUMN quality VARCHAR(50) DEFAULT 'common';

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_herbs_quality ON "herbs"(quality);
