// 心愿单概率提升配置
const wishlistBonus = {
  equipment: quality => 0.5, // 固定50%概率提升
  pet: rarity => 0.5 // 固定50%概率提升
}

// 获取调整后的装备概率
export const getAdjustedEquipProbabilities = (wishlistEnabled, selectedWishEquipQuality) => {
  const probabilities = {
    common: 0.4,
    uncommon: 0.3,
    rare: 0.15,
    epic: 0.1,
    legendary: 0.04,
    mythic: 0.01
  }

  // 如果启用了心愿单并且选择了特定品质，则对该品质进行概率提升
  if (wishlistEnabled && selectedWishEquipQuality) {
    const boost = wishlistBonus.equipment(selectedWishEquipQuality)
    const originalValue = probabilities[selectedWishEquipQuality]
    probabilities[selectedWishEquipQuality] = originalValue + originalValue * boost

    // 重新分配其他品质的概率
    const totalReduction = originalValue * boost
    const otherQualities = Object.keys(probabilities).filter(
      q => q !== selectedWishEquipQuality
    )
    const reductionPerQuality = totalReduction / otherQualities.length

    otherQualities.forEach(quality => {
      probabilities[quality] -= reductionPerQuality
    })
  }

  return probabilities
}

// 获取调整后的灵宠概率
export const getAdjustedPetProbabilities = (wishlistEnabled, selectedWishPetRarity) => {
  const probabilities = {
    common: 0.4,
    uncommon: 0.3,
    rare: 0.15,
    epic: 0.1,
    legendary: 0.04,
    mythic: 0.01
  }

  // 如果启用了心愿单并且选择了特定品质，则对该品质进行概率提升
  if (wishlistEnabled && selectedWishPetRarity) {
    const boost = wishlistBonus.pet(selectedWishPetRarity)
    const originalValue = probabilities[selectedWishPetRarity]
    probabilities[selectedWishPetRarity] = originalValue + originalValue * boost

    // 重新分配其他品质的概率
    const totalReduction = originalValue * boost
    const otherRarities = Object.keys(probabilities).filter(
      r => r !== selectedWishPetRarity
    )
    const reductionPerRarity = totalReduction / otherRarities.length

    otherRarities.forEach(rarity => {
      probabilities[rarity] -= reductionPerRarity
    })
  }

  return probabilities
}