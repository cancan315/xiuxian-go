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
    "spiritStones" INTEGER DEFAULT 0,
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
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL,
    "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Items table
CREATE TABLE IF NOT EXISTS "Items" (
    id SERIAL PRIMARY KEY,
    "userId" INTEGER REFERENCES "Users"(id) ON DELETE CASCADE,
    "itemId" VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    details JSON,
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL,
    "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Herbs table
CREATE TABLE IF NOT EXISTS "Herbs" (
    id SERIAL PRIMARY KEY,
    "userId" INTEGER REFERENCES "Users"(id) ON DELETE CASCADE,
    "herbId" VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    count INTEGER DEFAULT 1,
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL,
    "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Pills table
CREATE TABLE IF NOT EXISTS "Pills" (
    id SERIAL PRIMARY KEY,
    "userId" INTEGER REFERENCES "Users"(id) ON DELETE CASCADE,
    "pillId" VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    effect JSON,
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL,
    "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Artifacts table (equipment)
CREATE TABLE IF NOT EXISTS "Artifacts" (
    id SERIAL PRIMARY KEY,
    "userId" INTEGER REFERENCES "Users"(id) ON DELETE CASCADE,
    "artifactId" VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    slot VARCHAR(255),
    stats JSON,
    quality VARCHAR(255),
    equipped BOOLEAN DEFAULT false,
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL,
    "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_username ON "Users"(username);
CREATE INDEX IF NOT EXISTS idx_items_user_id ON "Items"("userId");
CREATE INDEX IF NOT EXISTS idx_herbs_user_id ON "Herbs"("userId");
CREATE INDEX IF NOT EXISTS idx_pills_user_id ON "Pills"("userId");
CREATE INDEX IF NOT EXISTS idx_artifacts_user_id ON "Artifacts"("userId");