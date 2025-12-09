-- Create tables for the xiuxian game

-- Users table
CREATE TABLE IF NOT EXISTS "Users" (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    "playerName" VARCHAR(255) DEFAULT '无名修士',
    level INTEGER DEFAULT 1,
    realm VARCHAR(255) DEFAULT '练气期一层',
    cultivation FLOAT DEFAULT 0,
    "maxCultivation" FLOAT DEFAULT 100,
    spirit FLOAT DEFAULT 0,
    "spiritStones" INTEGER DEFAULT 20000,
    "reinforceStones" INTEGER DEFAULT 0,
    "refinementStones" INTEGER DEFAULT 0,
    "baseAttributes" JSON DEFAULT '{"attack": 10, "health": 100, "defense": 5, "speed": 10}',
    "combatAttributes" JSON DEFAULT '{"critRate": 0, "comboRate": 0, "counterRate": 0, "stunRate": 0, "dodgeRate": 0, "vampireRate": 0}',
    "combatResistance" JSON DEFAULT '{"critResist": 0, "comboResist": 0, "counterResist": 0, "stunResist": 0, "dodgeResist": 0, "vampireResist": 0}',
    "specialAttributes" JSON DEFAULT '{"healBoost": 0, "critDamageBoost": 0, "critDamageReduce": 0, "finalDamageBoost": 0, "finalDamageReduce": 0, "combatBoost": 0, "resistanceBoost": 0}',
    "totalCultivationTime" INTEGER DEFAULT 0,
    "breakthroughCount" INTEGER DEFAULT 0,
    "explorationCount" INTEGER DEFAULT 0,
    "itemsFound" INTEGER DEFAULT 0,
    "eventTriggered" INTEGER DEFAULT 0,
    "isDarkMode" BOOLEAN DEFAULT false,
    "autoSellQualities" JSON DEFAULT '[]',
    "autoReleaseRarities" JSON DEFAULT '[]',
    "wishlistEnabled" BOOLEAN DEFAULT false,
    "selectedWishEquipQuality" VARCHAR(255),
    "selectedWishPetRarity" VARCHAR(255),
    "unlockedRealms" JSON DEFAULT '["练气一层"]',
    "unlockedLocations" JSON DEFAULT '["新手村"]',
    "unlockedSkills" JSON DEFAULT '[]',
    "isNewPlayer" BOOLEAN DEFAULT true,
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Items table (保留用于存储非装备类物品，如丹药、灵草等)
CREATE TABLE IF NOT EXISTS "Items" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "userId" INTEGER REFERENCES "Users"(id) ON DELETE CASCADE,
    "itemId" VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    details JSON,
    slot VARCHAR(255),
    stats JSON,
    quality VARCHAR(255),
    equipped BOOLEAN DEFAULT false,
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Equipment table (新建装备表)
CREATE TABLE IF NOT EXISTS "Equipment" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "userId" INTEGER REFERENCES "Users"(id) ON DELETE CASCADE,
    "equipmentId" VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    slot VARCHAR(255) NOT NULL,
    "equipType" VARCHAR(255),  -- 添加 equipType 字段用于映射抽奖系统中的 equipType
    details JSON,
    stats JSON,
    quality VARCHAR(255),
    "enhanceLevel" INTEGER DEFAULT 0,
    level INTEGER DEFAULT 1,  -- 添加 level 字段用于映射抽奖系统中的 level
    equipped BOOLEAN DEFAULT false,
    description VARCHAR(255),
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Pets table(新建灵宠表)
CREATE TABLE IF NOT EXISTS "Pets" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "userId" INTEGER REFERENCES "Users"(id) ON DELETE CASCADE,
    "petId" VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(255) DEFAULT 'pet',
    rarity VARCHAR(255) NOT NULL,
    level INTEGER DEFAULT 1,
    star INTEGER DEFAULT 0,
    experience INTEGER DEFAULT 0,
    "maxExperience" INTEGER DEFAULT 100,
    quality JSON,
    "combatAttributes" JSON,
    "isActive" BOOLEAN DEFAULT false,
    power INTEGER DEFAULT 0,
    "upgradeItems" INTEGER DEFAULT 1,
    description VARCHAR(255),
    "attackBonus" FLOAT DEFAULT 0,
    "defenseBonus" FLOAT DEFAULT 0,
    "healthBonus" FLOAT DEFAULT 0,
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Herbs table
CREATE TABLE IF NOT EXISTS "Herbs" (
    id SERIAL PRIMARY KEY,
    "userId" INTEGER REFERENCES "Users"(id) ON DELETE CASCADE,
    "herbId" VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    count INTEGER DEFAULT 1,
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Pills table
CREATE TABLE IF NOT EXISTS "Pills" (
    id SERIAL PRIMARY KEY,
    "userId" INTEGER REFERENCES "Users"(id) ON DELETE CASCADE,
    "pillId" VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    effect JSON,
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Artifacts table (removed - merged into Items table)

-- Add auto-update trigger for updated_at columns
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW."updatedAt" = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update the updatedAt column
DROP TRIGGER IF EXISTS update_users_updated_at ON "Users";
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON "Users" 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_items_updated_at ON "Items";
CREATE TRIGGER update_items_updated_at 
    BEFORE UPDATE ON "Items" 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_equipment_updated_at ON "Equipment";
CREATE TRIGGER update_equipment_updated_at 
    BEFORE UPDATE ON "Equipment" 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_pets_updated_at ON "Pets";
CREATE TRIGGER update_pets_updated_at 
    BEFORE UPDATE ON "Pets" 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_herbs_updated_at ON "Herbs";
CREATE TRIGGER update_herbs_updated_at 
    BEFORE UPDATE ON "Herbs" 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_pills_updated_at ON "Pills";
CREATE TRIGGER update_pills_updated_at 
    BEFORE UPDATE ON "Pills" 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Artifacts table triggers (removed)

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_username ON "Users"(username);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON "Users"("createdAt");
CREATE INDEX IF NOT EXISTS idx_items_user_id ON "Items"("userId");
CREATE INDEX IF NOT EXISTS idx_equipment_user_id ON "Equipment"("userId");
CREATE INDEX IF NOT EXISTS idx_pets_user_id ON "Pets"("userId");
CREATE INDEX IF NOT EXISTS idx_pets_active ON "Pets"("isActive");
CREATE INDEX IF NOT EXISTS idx_herbs_user_id ON "Herbs"("userId");
CREATE INDEX IF NOT EXISTS idx_pills_user_id ON "Pills"("userId");
-- Artifact indexes (removed)