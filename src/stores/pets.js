import { defineStore } from 'pinia'
import { itemQualities } from '../plugins/itemQualities'

export const usePetsStore = defineStore('pets', {
  state: () => ({
    petConfig: {
      rarityMap: itemQualities.pet
    },
    pets: [], // 灵宠库存
  }),
})