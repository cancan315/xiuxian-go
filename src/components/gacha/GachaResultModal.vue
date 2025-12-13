<template>
  <n-modal
    v-model:show="show"
    preset="dialog"
    title="抽卡结果"
    :style="{ maxWidth: '90vw', width: '800px' }"
  >
    <n-card :bordered="false">
      <!-- 筛选区域 -->
      <div class="filter-section" v-if="gachaType !== 'all'">
        <n-space align="center" justify="center" :wrap="true" :size="16">
          <n-select
            v-model:value="localSelectedQuality"
            placeholder="装备品质筛选"
            clearable
            :options="equipmentQualityOptions"
            :style="{ width: '180px' }"
            @update:value="handleQualityChange"
            v-if="gachaType === 'equipment'"
          ></n-select>
          <n-select
            v-model:value="localSelectedRarity"
            placeholder="灵宠品质筛选"
            clearable
            :options="petRarityOptions"
            :style="{ width: '180px' }"
            @update:value="handleRarityChange"
            v-if="gachaType === 'pet'"
          ></n-select>
        </n-space>
      </div>
      <n-space justify="center">
        <n-button
          type="primary"
          @click="handlePerformGacha(gachaNumber)"
          :disabled="
            spiritStones < (wishlistEnabled ? gachaNumber * 200 : gachaNumber * 100) ||
            isDrawing
          "
        >
          再抽{{ gachaNumber }}次 ({{
            wishlistEnabled ? gachaNumber * 200 : gachaNumber * 100
          }}灵石)
        </n-button>
      </n-space>
      
      <ResultGrid 
        :current-page-results="gachaStore.currentPageResults"
        :wishlist-enabled="wishlistEnabled"
        :selected-wish-equip-quality="selectedWishEquipQuality"
        :selected-wish-pet-rarity="selectedWishPetRarity"
        :pet-rarities="petRarities"
        :equipment-types-map="equipmentTypesMap"
        @show-pet-detail="showPetDetail"
        @show-equipment-detail="showEquipmentDetail"
      />
      
      <template #footer>
        <n-space justify="center">
          <n-pagination
            v-model:page="currentPage"
            :page-slot="6"
            :page-count="gachaStore.totalPages"
            :page-size="pageSize"
          />
        </n-space>
      </template>
    </n-card>
  </n-modal>
</template>

<script setup>
  import { ref, computed, defineProps, defineEmits } from 'vue'
  import ResultGrid from './ResultGrid.vue'
  import { useGachaStore } from '../../stores/gacha'

  const gachaStore = useGachaStore()

  const props = defineProps({
    show: {
      type: Boolean,
      default: false
    },
    gachaType: {
      type: String,
      required: true
    },
    gachaNumber: {
      type: Number,
      required: true
    },
    spiritStones: {
      type: Number,
      required: true
    },
    wishlistEnabled: {
      type: Boolean,
      default: false
    },
    isDrawing: {
      type: Boolean,
      default: false
    },
    currentPageResults: {
      type: Array,
      required: true
    },
    totalPages: {
      type: Number,
      required: true
    },
    pageSize: {
      type: Number,
      required: true
    },
    currentPage: {
      type: Number,
      required: true
    },
    selectedWishEquipQuality: {
      type: String,
      default: ''
    },
    selectedWishPetRarity: {
      type: String,
      default: ''
    },
    petRarities: {
      type: Object,
      required: true
    },
    equipmentQualityOptions: {
      type: Array,
      required: true
    },
    petRarityOptions: {
      type: Array,
      required: true
    },
    equipmentTypesMap: {
      type: Object,
      required: true
    }
  })

  const emit = defineEmits([
    'update:show',
    'update:current-page',
    'perform-gacha',
    'show-pet-detail',
    'show-equipment-detail'
  ])

  const localSelectedQuality = ref(props.gachaType === 'equipment' ? '' : 'all')
  const localSelectedRarity = ref(props.gachaType === 'pet' ? '' : 'all')

  const show = computed({
    get: () => props.show,
    set: (value) => emit('update:show', value)
  })

  const currentPage = computed({
    get: () => props.currentPage,
    set: (value) => emit('update:current-page', value)
  })

  const handleQualityChange = (value) => {
    localSelectedQuality.value = value
    currentPage.value = 1
    // 同步到 store
    gachaStore.setSelectedQuality(value)
  }

  const handleRarityChange = (value) => {
    localSelectedRarity.value = value
    currentPage.value = 1
    // 同步到 store
    gachaStore.setSelectedRarity(value)
  }

  const handlePerformGacha = (count) => {
    emit('perform-gacha', count)
  }

  const showPetDetail = (pet) => {
    emit('show-pet-detail', pet)
  }

  const showEquipmentDetail = (equipment) => {
    emit('show-equipment-detail', equipment)
  }
</script>

<style scoped>
  /* 样式已在 ResultGrid 组件中定义 */
</style>