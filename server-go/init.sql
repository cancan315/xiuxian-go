-- 创建数据库表结构，基于Go后端模型定义
-- 表名命名规范：复数名词  --
-- users
-- items
-- user_items  -- 关联表命名规范用下划线连接

-- 字段名命名规范：小写+下划线

-- users 表
CREATE TABLE IF NOT EXISTS "users" (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    player_name VARCHAR(255),
    level INTEGER DEFAULT 1,
    realm VARCHAR(255),
    cultivation DOUBLE PRECISION DEFAULT 0,
    max_cultivation DOUBLE PRECISION DEFAULT 100,
    spirit DOUBLE PRECISION DEFAULT 0,
    spirit_stones INTEGER DEFAULT 0,
    reinforce_stones INTEGER DEFAULT 0,
    refinement_stones INTEGER DEFAULT 0,
    pet_essence INTEGER DEFAULT 0,
    name_change_count INTEGER DEFAULT 0,
    base_attributes JSONB,
    combat_attributes JSONB,
    combat_resistance JSONB,
    special_attributes JSONB,
    last_spirit_gain_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- items 表
CREATE TABLE IF NOT EXISTS "items" (
    id UUID PRIMARY KEY,
    user_id INTEGER REFERENCES "users"(id),
    item_id VARCHAR(255),
    name VARCHAR(255),
    type VARCHAR(255),
    details JSONB,
    slot VARCHAR(255),
    stats JSONB,
    quality VARCHAR(50),
    equipped BOOLEAN DEFAULT FALSE
);

-- herbs 表
CREATE TABLE IF NOT EXISTS "herbs" (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES "users"(id),
    herb_id VARCHAR(255),
    name VARCHAR(255),
    count INTEGER DEFAULT 0,
    quality VARCHAR(50) DEFAULT 'common'
);

-- pills 表
CREATE TABLE IF NOT EXISTS "pills" (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES "users"(id),
    pill_id VARCHAR(255),
    name VARCHAR(255),
    description TEXT,
    effect JSONB
);

-- pill_fragments 表
CREATE TABLE IF NOT EXISTS "pill_fragments" (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES "users"(id),
    recipe_id VARCHAR(255),
    count INTEGER DEFAULT 0
);

-- user_alchemy_data 表
CREATE TABLE IF NOT EXISTS "user_alchemy_data" (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE REFERENCES "users"(id),
    recipes_unlocked JSONB,  -- 已解锁的丹方ID列表
    pills_crafted INTEGER DEFAULT 0,   -- 总炼制次数
    pills_consumed INTEGER DEFAULT 0,  -- 总服用次数
    alchemy_level INTEGER DEFAULT 1,   -- 炼丹等级
    alchemy_rate DOUBLE PRECISION DEFAULT 1.0  -- 炼丹加成率
);

-- pets 表
CREATE TABLE IF NOT EXISTS "pets" (
    id UUID PRIMARY KEY,
    user_id INTEGER REFERENCES "users"(id),
    pet_id VARCHAR(255),
    name VARCHAR(255),
    type VARCHAR(255),
    rarity VARCHAR(50),
    level INTEGER DEFAULT 1,
    star INTEGER DEFAULT 1,
    experience INTEGER DEFAULT 0,
    max_experience INTEGER DEFAULT 100,
    quality JSONB,
    combat_attributes JSONB,
    is_active BOOLEAN DEFAULT FALSE,
    attack_bonus DOUBLE PRECISION DEFAULT 0,
    defense_bonus DOUBLE PRECISION DEFAULT 0,
    health_bonus DOUBLE PRECISION DEFAULT 0
);

-- equipment 表
CREATE TABLE IF NOT EXISTS "equipment" (
    id UUID PRIMARY KEY,
    user_id INTEGER REFERENCES "users"(id),
    equipment_id VARCHAR(255),
    name VARCHAR(255),
    type VARCHAR(255),
    slot VARCHAR(255),
    equip_type VARCHAR(255),
    details JSONB,
    stats JSONB,
    extra_attributes JSONB,
    quality VARCHAR(50),
    enhance_level INTEGER DEFAULT 0,
    equipped BOOLEAN DEFAULT FALSE,
    description TEXT,
    required_realm INTEGER DEFAULT 1,
    level INTEGER DEFAULT 1
);

-- dungeon_progress 表 (秘境进度追踪)
CREATE TABLE IF NOT EXISTS "dungeon_progress" (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES "users"(id) ON DELETE CASCADE,
    current_floor INTEGER DEFAULT 1,
    max_floor_reached INTEGER DEFAULT 1,
    current_difficulty VARCHAR(50),
    total_runs INTEGER DEFAULT 0,
    total_victories INTEGER DEFAULT 0,
    total_losses INTEGER DEFAULT 0,
    total_rewards_earned JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

-- dungeon_buffs 表 (翔义选择记录)
CREATE TABLE IF NOT EXISTS "dungeon_buffs" (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES "users"(id) ON DELETE CASCADE,
    dungeon_run_id VARCHAR(255),
    floor INTEGER,
    buff_id VARCHAR(255),
    buff_name VARCHAR(255),
    buff_type VARCHAR(50),
    buff_effects JSONB,
    selected_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- battle_records 表 (斗法战斗记录)
CREATE TABLE IF NOT EXISTS "battle_records" (
    id SERIAL PRIMARY KEY,
    player_id INTEGER NOT NULL REFERENCES "users"(id) ON DELETE CASCADE,
    opponent_id INTEGER NOT NULL REFERENCES "users"(id) ON DELETE CASCADE,
    opponent_name VARCHAR(255),
    result VARCHAR(50) NOT NULL,
    battle_type VARCHAR(50) NOT NULL,
    rewards VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_users_username ON "users"(username);
CREATE INDEX IF NOT EXISTS idx_users_last_spirit_gain_time ON "users"(last_spirit_gain_time);
CREATE INDEX IF NOT EXISTS idx_items_user_id ON "items"(user_id);
CREATE INDEX IF NOT EXISTS idx_herbs_user_id ON "herbs"(user_id);
CREATE INDEX IF NOT EXISTS idx_herbs_quality ON "herbs"(quality);
CREATE INDEX IF NOT EXISTS idx_pills_user_id ON "pills"(user_id);
CREATE INDEX IF NOT EXISTS idx_pill_fragments_user_id ON "pill_fragments"(user_id);
CREATE INDEX IF NOT EXISTS idx_user_alchemy_data_user_id ON "user_alchemy_data"(user_id);
CREATE INDEX IF NOT EXISTS idx_pets_user_id ON "pets"(user_id);
CREATE INDEX IF NOT EXISTS idx_equipment_user_id ON "equipment"(user_id);
CREATE INDEX IF NOT EXISTS idx_dungeon_progress_user_id ON "dungeon_progress"(user_id);
CREATE INDEX IF NOT EXISTS idx_dungeon_buffs_user_id ON "dungeon_buffs"(user_id);
CREATE INDEX IF NOT EXISTS idx_dungeon_buffs_run_id ON "dungeon_buffs"(dungeon_run_id);
CREATE INDEX IF NOT EXISTS idx_battle_records_player_id ON "battle_records"(player_id);
CREATE INDEX IF NOT EXISTS idx_battle_records_opponent_id ON "battle_records"(opponent_id);
CREATE INDEX IF NOT EXISTS idx_battle_records_created_at ON "battle_records"(created_at);
