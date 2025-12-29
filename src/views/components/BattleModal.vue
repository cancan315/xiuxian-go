<template>
  <!-- 战斗模拟弹窗 -->
  <n-modal 
    v-model:show="show" 
    preset="card"
    :title="title"
    style="width: 90%; max-width: 800px;"
    :bordered="false"
    size="huge"
  >
    <div v-if="battleLogs.length > 0">
      <!-- 战斗日志滚动容器 -->
      <n-scrollbar style="max-height: 400px; margin-bottom: 16px;">
        <div class="battle-log-container">
          <!-- 战斗日志项 -->
          <div v-for="(log, index) in battleLogs" :key="index" class="battle-log-item">
            <n-tag :type="getLogTagType(log.type)" size="small" style="margin-right: 8px;">
              {{ log.type }}
            </n-tag>
            <span>{{ log.message }}</span>
          </div>
        </div>
      </n-scrollbar>
      
      <!-- 战斗结果展示 -->
      <div class="battle-result" v-if="battleResult">
        <n-alert :title="battleResult.title" :type="battleResult.type">
          <template #icon>
            <n-icon :component="battleResult.icon" />
          </template>
          {{ battleResult.message }}
          
          <!-- 战斗奖励展示 -->
          <template v-if="battleResult.rewards && battleResult.rewards.length > 0">
            <n-divider style="margin: 8px 0;" />
            <div>获得奖励：</div>
            <n-space style="margin-top: 8px;">
              <n-tag v-for="(reward, idx) in battleResult.rewards" :key="idx" type="success">
                {{ reward }}
              </n-tag>
            </n-space>
          </template>
        </n-alert>
      </div>

      <!-- 战斗弹窗操作按钮 -->
      <n-space justify="end" style="margin-top: 16px;">
        <n-button @click="$emit('close')">关闭</n-button>
        <n-button type="primary" v-if="battleResult && battleResult.win" @click="$emit('claim-rewards')">
          领取奖励
        </n-button>
      </n-space>
    </div>
    <!-- 战斗加载状态 -->
    <div v-else>
      <n-space justify="center" style="padding: 40px;">
        <n-spin size="large" />
      </n-space>
    </div>
  </n-modal>
</template>

<script setup>
import { computed } from 'vue'
import { NModal, NScrollbar, NTag, NAlert, NDivider, NSpace, NButton, NIcon, NSpin } from 'naive-ui'
import { getLogTagType } from '../utils/duelHelper'

// 定义 props
const props = defineProps({
  // 是否显示模态框
  show: {
    type: Boolean,
    default: false
  },
  // 战斗标题
  title: {
    type: String,
    default: '战斗模拟'
  },
  // 战斗日志
  battleLogs: {
    type: Array,
    default: () => []
  },
  // 战斗结果
  battleResult: {
    type: Object,
    default: null
  }
})

// 定义 emit 事件
const emit = defineEmits(['update:show', 'close', 'claim-rewards'])

// 处理 show 属性变化
const show = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value)
})
</script>

<style scoped>
.battle-log-container {
  background-color: rgba(0, 0, 0, 0.02);
  border-radius: 8px;
  padding: 12px;
}

.dark .battle-log-container {
  background-color: rgba(255, 255, 255, 0.02);
}

.battle-log-item {
  padding: 6px 0;
  border-bottom: 1px solid #eee;
  font-family: 'Courier New', monospace;
  font-size: 14px;
}

.dark .battle-log-item {
  border-bottom: 1px solid #444;
}

.battle-log-item:last-child {
  border-bottom: none;
}

.battle-result {
  margin-top: 16px;
  padding: 16px;
  background-color: rgba(0, 0, 0, 0.02);
  border-radius: 8px;
}

.dark .battle-result {
  background-color: rgba(255, 255, 255, 0.02);
}
</style>
