<template>
  <div class="result-grid">
    <div
      v-for="item in currentPageResults"
      :key="item.id"
      :class="[
        'result-item',
        {
          'wish-bonus':
            wishlistEnabled &&
            ((item.qualityInfo && selectedWishEquipQuality === item.quality) ||
              (item.type === 'pet' && selectedWishPetRarity === item.rarity))
        }
      ]"
      :style="{
        borderColor: item.qualityInfo
          ? item.qualityInfo.color
          : petRarities[item.rarity]?.color || '#CCCCCC'
      }"
    >
      <h4>{{ item.name }}</h4>
      <p>品质：{{ item.qualityInfo ? item.qualityInfo.name : petRarities[item.rarity]?.name || '未知' }}</p>
      <p v-if="equipmentTypes.includes(item.type)">类型：{{ equipmentTypesMap[item.equipType]?.name }}</p>
      <p v-else-if="item.type === 'pet'">{{ item.description || '暂无描述' }}</p>
      <!-- 添加详情按钮 -->
      <n-button quaternary circle size="small" @click="showPetDetail(item)" v-if="item.type === 'pet'">
        <template #icon>
          <n-icon>
            <InformationOutline /> 
          </n-icon>
        </template>
      </n-button>
      <n-button quaternary circle size="small" @click="showEquipmentDetail(item)" v-if="item.type === 'equipment'">
        <template #icon>
          <n-icon>
            <InformationOutline /> 
          </n-icon>
        </template>
      </n-button>
    </div>
  </div>
</template>

<script setup>
  import { defineProps, defineEmits } from 'vue'
  import { InformationOutline } from '@vicons/ionicons5'

  const props = defineProps({
    currentPageResults: {
      type: Array,
      required: true
    },
    wishlistEnabled: {
      type: Boolean,
      default: false
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
    equipmentTypesMap: {
      type: Object,
      required: true
    }
  })

  const equipmentTypes = ['weapon', 'armor', 'accessory']

  const emit = defineEmits(['show-pet-detail', 'show-equipment-detail'])

  const showPetDetail = (pet) => {
    emit('show-pet-detail', pet)
  }

  const showEquipmentDetail = (equipment) => {
    emit('show-equipment-detail', equipment)
  }
</script>

<style scoped>
  .result-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 16px;
    margin: 20px 0;
  }

  .result-item {
    border: 2px solid #ccc;
    border-radius: 8px;
    padding: 12px;
    text-align: center;
    background-color: #f9f9f9;
    position: relative;
  }

  .result-item h4 {
    margin: 0 0 8px 0;
  }

  .result-item p {
    margin: 4px 0;
    font-size: 0.9em;
  }

  @media screen and (max-width: 768px) {
    .result-grid {
      grid-template-columns: repeat(2, 1fr);
    }
  }

  .wish-bonus {
    position: relative;
    z-index: 1;
  }

  .wish-bonus::before {
    content: '★';
    position: absolute;
    top: -10px;
    right: -10px;
    color: white;
    font-size: 20px;
    text-shadow: 0 0 5px;
    animation: rotate-stars 3s linear infinite;
    transform-origin: center;
  }

  @keyframes rotate-stars {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }
</style>