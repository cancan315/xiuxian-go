<template>
  <div class="gacha-item-container">
    <div
      class="gacha-item"
      :class="{
        shake: isShaking,
        open: isOpening
      }"
    >
      {{ types[gachaType] }}
    </div>
  </div>
</template>

<script setup>
  import { defineProps, defineEmits } from 'vue'

  const props = defineProps({
    gachaType: {
      type: String,
      required: true
    },
    isShaking: {
      type: Boolean,
      default: false
    },
    isOpening: {
      type: Boolean,
      default: false
    }
  })

  const types = {
    equipment: '🥋',
    pet: '🥚'
  }
</script>

<style scoped>
  .gacha-item-container {
    margin: 30px 0;
    display: flex;
    justify-content: center;
  }

  .gacha-item {
    width: 200px;
    height: 200px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-size: 80px;
    font-weight: bold;
    transition: transform 0.3s ease;
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.3);
  }

  .gacha-item.shake {
    animation: shake 0.5s ease-in-out infinite;
  }

  .gacha-item.open {
    transform: scale(1.1);
    background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  }

  @keyframes shake {
    0% {
      transform: translateX(0);
    }
    25% {
      transform: translateX(-5px);
    }
    50% {
      transform: translateX(5px);
    }
    75% {
      transform: translateX(-5px);
    }
    100% {
      transform: translateX(0);
    }
  }
</style>