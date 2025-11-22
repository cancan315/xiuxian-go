<template>
  <n-layout>
    <n-layout-header bordered>
      <n-page-header>
        <template #title>图鉴</template>
      </n-page-header>
    </n-layout-header>
    <n-layout-content class="gallery-content">
      <n-card :bordered="false">
        <n-result status="404" title="功能开发中" description="图鉴功能正在开发中，敬请期待">
          <template #footer>
            <n-button @click="router.push('/cultivation')">返回修炼</n-button>
          </template>
        </n-result>
      </n-card>
    </n-layout-content>
  </n-layout>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<style scoped>
.gallery-content {
  padding: 16px;
}
</style>