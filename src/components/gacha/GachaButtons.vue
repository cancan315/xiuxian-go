<template>
  <div class="gacha-buttons">
    <n-space vertical>
      <n-space justify="center">
        <n-button
          type="primary"
          v-for="(item, index) in [1, 10, 50, 100]"
          :key="index"
          @click="handleGacha(item)"
          :disabled="
            spiritStones < (wishlistEnabled ? item * 200 : item * 100) || isDrawing
          "
        >
          抽{{ item }}次 ({{ wishlistEnabled ? item * 200 : item * 100 }}灵石)
        </n-button>
      </n-space>
      <n-space justify="center">
        <n-button quaternary circle size="small" @click="showProbability">
          <template #icon>
            <n-icon>
              <Help />
            </n-icon>
          </template>
        </n-button>
        <n-button quaternary circle size="small" @click="showWishlist">
          <template #icon>
            <n-icon>
              <HeartOutline />
            </n-icon>
          </template>
        </n-button>
        <n-button quaternary circle size="small" @click="showAutoSettings">
          <template #icon>
            <n-icon>
              <SettingsOutline />
            </n-icon>
          </template>
        </n-button>
      </n-space>
    </n-space>
  </div>
</template>

<script setup>
  import { defineProps, defineEmits } from 'vue'
  import { Help, HeartOutline, SettingsOutline } from '@vicons/ionicons5'

  const props = defineProps({
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
    }
  })

  const emit = defineEmits(['gacha', 'show-probability', 'show-wishlist', 'show-auto-settings'])

  const handleGacha = (count) => {
    emit('gacha', count)
  }

  const showProbability = () => {
    emit('show-probability')
  }

  const showWishlist = () => {
    emit('show-wishlist')
  }

  const showAutoSettings = () => {
    emit('show-auto-settings')
  }
</script>

<style scoped>
  .gacha-buttons {
    margin-top: 20px;
  }
</style>