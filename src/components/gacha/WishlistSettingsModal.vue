<template>
  <n-modal v-model:show="show" preset="dialog" title="心愿单设置" style="width: 800px">
    <n-card :bordered="false">
      <n-space vertical>
        <n-switch v-model:value="localWishlistEnabled">
          <template #checked>心愿单已启用</template>
          <template #unchecked>心愿单已禁用</template>
        </n-switch>
        <n-divider>装备品质心愿</n-divider>
        <n-select
          v-model:value="localSelectedWishEquipQuality"
          :options="equipmentQualityOptions"
          clearable
          placeholder="选择装备品质"
          :disabled="!localWishlistEnabled"
        >
          <template #option="{ option }">
            <span :style="{ color: equipmentQualities[option.value].color }">
              {{ equipmentQualities[option.value].name }}
              <n-tag v-if="option.value === localSelectedWishEquipQuality" type="success" size="small">
                已选择
              </n-tag>
            </span>
          </template>
        </n-select>
        <n-divider>灵宠品质心愿</n-divider>
        <n-select
          v-model:value="localSelectedWishPetRarity"
          :options="petRarityOptions"
          clearable
          placeholder="选择灵宠品质"
          :disabled="!localWishlistEnabled"
        >
          <template #option="{ option }">
            <span :style="{ color: petRarities[option.value].color }">
              {{ petRarities[option.value].name }}
              <n-tag v-if="option.value === localSelectedWishPetRarity" type="success" size="small">
                已选择
              </n-tag>
            </span>
          </template>
        </n-select>
        <n-alert type="info" title="心愿单说明">
          启用心愿单后，所需灵石会翻倍,
          选中的品质将获得50%的概率提升。每次只能选择一个装备品质和一个灵宠品质作为心愿。
        </n-alert>
      </n-space>
    </n-card>
  </n-modal>
</template>

<script setup>
  import { ref, computed, defineProps, defineEmits } from 'vue'

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
    equipmentQualityOptions: {
      type: Array,
      required: true
    },
    petRarityOptions: {
      type: Array,
      required: true
    }
  })

  const emit = defineEmits(['update:show', 'update:wishlist-enabled', 'update:selected-wish-equip-quality', 'update:selected-wish-pet-rarity'])

  const show = computed({
    get: () => props.show,
    set: (value) => emit('update:show', value)
  })

  const localWishlistEnabled = computed({
    get: () => props.wishlistEnabled,
    set: (value) => emit('update:wishlist-enabled', value)
  })

  const localSelectedWishEquipQuality = computed({
    get: () => props.selectedWishEquipQuality,
    set: (value) => emit('update:selected-wish-equip-quality', value)
  })

  const localSelectedWishPetRarity = computed({
    get: () => props.selectedWishPetRarity,
    set: (value) => emit('update:selected-wish-pet-rarity', value)
  })
</script>