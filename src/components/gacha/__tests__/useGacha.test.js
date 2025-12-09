import { describe, it, expect } from 'vitest'
import { generateRandomEquipment, generateRandomPet } from '../useGacha'

describe('useGacha', () => {
  describe('generateRandomEquipment', () => {
    it('generates equipment with correct structure', () => {
      const equipment = generateRandomEquipment(10)
      
      // 检查必需的属性
      expect(equipment).toHaveProperty('id')
      expect(equipment).toHaveProperty('name')
      expect(equipment).toHaveProperty('type', 'equipment')
      expect(equipment).toHaveProperty('quality')
      expect(equipment).toHaveProperty('qualityInfo')
      expect(equipment).toHaveProperty('equipType')
      expect(equipment).toHaveProperty('level')
      expect(equipment).toHaveProperty('stats')
      expect(equipment).toHaveProperty('extraAttributes')
      expect(equipment).toHaveProperty('createdAt')
      
      // 检查stats结构
      expect(equipment.stats).toHaveProperty('attack')
      expect(equipment.stats).toHaveProperty('defense')
      expect(equipment.stats).toHaveProperty('health')
      expect(equipment.stats).toHaveProperty('speed')
    })

    it('generates equipment with valid level based on player level', () => {
      const equipment = generateRandomEquipment(10)
      expect(equipment.level).toBeGreaterThanOrEqual(1)
    })
  })

  describe('generateRandomPet', () => {
    it('generates pet with correct structure', () => {
      const pet = generateRandomPet(10)
      
      // 检查必需的属性
      expect(pet).toHaveProperty('id')
      expect(pet).toHaveProperty('name')
      expect(pet).toHaveProperty('type', 'pet')
      expect(pet).toHaveProperty('rarity')
      expect(pet).toHaveProperty('level')
      expect(pet).toHaveProperty('star')
      expect(pet).toHaveProperty('exp')
      expect(pet).toHaveProperty('description')
      expect(pet).toHaveProperty('combatAttributes')
      expect(pet).toHaveProperty('createdAt')
      
      // 检查combatAttributes结构
      expect(pet.combatAttributes).toHaveProperty('attack')
      expect(pet.combatAttributes).toHaveProperty('defense')
      expect(pet.combatAttributes).toHaveProperty('health')
      expect(pet.combatAttributes).toHaveProperty('speed')
    })

    it('generates pet with level 1', () => {
      const pet = generateRandomPet(10)
      expect(pet.level).toBe(1)
    })
  })
})