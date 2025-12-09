<template>
  <n-modal v-model:show="show" preset="dialog" title="抽卡概率说明">
    <n-tabs type="segment" animated>
      <!-- 装备池概率 -->
      <n-tab-pane name="equipment" tab="装备池">
        <n-card>
          <div class="probability-bars">
            <div
              v-for="(probability, quality) in adjustedEquipProbabilities"
              :key="quality"
              class="prob-item"
            >
              <div class="prob-label">
                <span :style="{ color: equipmentQualities[quality].color }">
                  {{ equipmentQualities[quality].name }}
                </span>
              </div>
              <n-progress
                type="line"
                :percentage="probability * 100"
                :indicator-placement="'inside'"
                :color="equipmentQualities[quality].color"
                :height="20"
                :border-radius="4"
                :class="{
                  'wish-bonus': wishlistEnabled && selectedWishEquipQuality === quality
                }"
                :show-indicator="true"
              >
                <template #indicator>{{ (probability * 100).toFixed(1) }}%</template>
              </n-progress>
            </div>
          </div>
        </n-card>
      </n-tab-pane>
      <!-- 灵宠池概率 -->
      <n-tab-pane name="pet" tab="灵宠池">
        <n-card>
          <div class="probability-bars">
            <div v-for="(probability, rarity) in adjustedPetProbabilities" :key="rarity" class="prob-item">
              <div class="prob-label">
                <span :style="{ color: petRarities[rarity].color }">
                  {{ petRarities[rarity].name }}
                </span>
              </div>
              <n-progress
                type="line"
                :percentage="probability * 100"
                :indicator-placement="'inside'"
                :class="{
                  'wish-bonus': wishlistEnabled && selectedWishPetRarity === rarity
                }"
                :color="petRarities[rarity].color"
                :height="20"
                :border-radius="4"
                :show-indicator="true"
              >
                <template #indicator>{{ (probability * 100).toFixed(1) }}%</template>
              </n-progress>
            </div>
          </div>
        </n-card>
      </n-tab-pane>
    </n-tabs>
  </n-modal>
</template>

<script setup>
  import { computed, defineProps, defineEmits } from 'vue'

  const props = defineProps({
    show: {
      type: Boolean,
      default: false
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
    equipmentQualities: {
      type: Object,
      required: true
    },
    petRarities: {
      type: Object,
      required: true
    },
    adjustedEquipProbabilities: {
      type: Object,
      required: true
    },
    adjustedPetProbabilities: {
      type: Object,
      required: true
    }
  })

  const emit = defineEmits(['update:show'])

  const show = computed({
    get: () => props.show,
    set: (value) => emit('update:show', value)
  })
</script>

<style scoped>
  .probability-bars {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .prob-item {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .prob-label {
    min-width: 60px;
    text-align: right;
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