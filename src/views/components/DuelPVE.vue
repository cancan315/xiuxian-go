<template>
  <div class="pve-section">
    <!-- 妖兽挑战说明 -->
    <n-alert title="妖兽挑战--没开发完，别试" type="warning" style="margin-bottom: 16px;">
      挑战不同等级的妖兽，获得妖兽材料和修为奖励。
    </n-alert>
    
    <!-- 妖兽难度选择 -->
    <n-card title="选择挑战难度" size="small">
      <n-space vertical>
        <!-- 难度选择单选组 -->
        <n-radio-group v-model:value="selectedDifficulty" name="difficulty">
          <n-space>
            <n-radio v-for="difficulty in difficulties" :key="difficulty.value" :value="difficulty.value">
              {{ difficulty.label }}
            </n-radio>
          </n-space>
        </n-radio-group>
        
        <!-- 妖兽列表 -->
        <n-list bordered>
          <n-list-item v-for="monster in filteredMonsters" :key="monster.id">
            <n-thing>
              <template #header>
                <n-space align="center">
                  <span>{{ monster.name }}</span>
                  <!-- 难度标签 -->
                  <n-tag :type="getDifficultyTagType(monster.difficulty)">
                    {{ getDifficultyName(monster.difficulty) }}
                  </n-tag>
                </n-space>
              </template>
              <template #description>
                <!-- 妖兽属性描述 -->
                <n-descriptions label-placement="left" :column="2" size="small">
                  <n-descriptions-item label="血量">{{ monster.health }}</n-descriptions-item>
                  <n-descriptions-item label="攻击">{{ monster.attack }}</n-descriptions-item>
                  <n-descriptions-item label="防御">{{ monster.defense }}</n-descriptions-item>
                  <n-descriptions-item label="速度">{{ monster.speed }}</n-descriptions-item>
                </n-descriptions>
                <div style="margin-top: 8px;">
                  掉落: {{ monster.rewards }}
                </div>
              </template>
              <template #footer>
                <n-space justify="end">
                  <!-- 挑战妖兽按钮 -->
                  <n-button type="primary" size="small" @click="$emit('challenge-monster', monster)">
                    降服
                  </n-button>
                  <!-- 查看妖兽详细信息按钮 -->
                  <n-button size="small" @click="$emit('view-monster-info', monster)">
                    详细信息
                  </n-button>
                </n-space>
              </template>
            </n-thing>
          </n-list-item>
        </n-list>
      </n-space>
    </n-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { 
  NCard, NAlert, NSpace, NButton, NList, NListItem, NThing, NTag, 
  NDescriptions, NDescriptionsItem, NRadioGroup, NRadio 
} from 'naive-ui'
import APIService from '../../services/api'
import { getAuthToken } from '../../stores/db'
import { getDifficultyTagType, getDifficultyName } from '../utils/duelHelper'

// 定义 emit 事件
const emit = defineEmits(['challenge-monster', 'view-monster-info'])

// 状态管理
const selectedDifficulty = ref('normal')
const monsters = ref([])

// 难度选项
const difficulties = [
  { label: '普通', value: 'normal' },
  { label: '困难', value: 'hard' },
  { label: '噩梦', value: 'nightmare' }
]

/**
 * 过滤后的妖兽列表，根据选中的难度过滤
 */
const filteredMonsters = computed(() => {
  return monsters.value.filter(monster => 
    selectedDifficulty.value === 'all' || monster.difficulty === selectedDifficulty.value
  )
})

/**
 * 加载妖兽列表
 */
const loadMonsters = async () => {
  try {
    const token = getAuthToken()
    const response = await APIService.getMonsters(token)
    if (response.success) {
      monsters.value = response.data.monsters
    }
  } catch (error) {
    console.error('获取妖兽列表失败:', error)
  }
}

// 初始化加载
onMounted(() => {
  loadMonsters()
})
</script>

<style scoped>
.pve-section {
  padding: 8px;
}
</style>
