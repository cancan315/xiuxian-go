<template>
  <div class="settings-container">
    <n-card title="游戏设置">
      <template #header-extra>游戏版本{{ version }}</template>
      <n-space vertical>
        <n-input-group>
          <n-input v-model:value="newName" placeholder="输入新的道号" clearable :maxlength="maxLength" show-count />
          <n-button type="primary" @click="handleChangeName" :disabled="!newName">修改道号</n-button>
        </n-input-group>
        <n-alert title="必看说明" type="warning">
          本游戏免费，如果您在任何地方通过付费方式购买了本游戏，请及时退款并投诉举报。如需赞助，请联系群主。
        </n-alert>
        <n-space>
          <n-button type="error" @click="qq = true">官方群聊</n-button>
        </n-space>
      </n-space>
    </n-card>
    <n-modal preset="dialog" title="玩家交流群" v-model:show="qq">
      <n-card :bordered="false" size="huge" role="dialog" aria-modal="true">
        <n-space vertical>
          <n-input value="755301571" readonly type="text" />
        </n-space>
      </n-card>
    </n-modal>
  </div>
</template>

<script setup>
  import { usePlayerStore } from '../stores/player'
  import { ref } from 'vue'
  import { useDialog, useMessage } from 'naive-ui'

  const newName = ref('')
  const message = useMessage()
  const maxLength = 6 // 定义道号最大长度常量
  const playerStore = usePlayerStore()
  const dialog = useDialog()
  const version = __APP_VERSION__

  const qq = ref(false)

  // 修改道号
  const handleChangeName = () => {
    if (!newName.value.trim()) {
      message.warning('道号不能为空！')
      return
    }
    if (newName.value.trim().length > maxLength) {
      message.warning(`道号长度不能超过${maxLength}个字符！`)
      return
    }
    // 计算修改道号所需灵石
    const spiritStoneCost = playerStore.nameChangeCount === 0 ? 0 : Math.pow(2, playerStore.nameChangeCount) * 100
    // 第一次修改免费，之后需要消耗灵石
    if (playerStore.nameChangeCount > 0) {
      if (playerStore.spiritStones < spiritStoneCost) {
        message.error(`灵石不足！修改道号需要${spiritStoneCost}颗灵石`)
        return
      }
      playerStore.spiritStones -= spiritStoneCost
    }
    playerStore.name = newName.value.trim()
    playerStore.nameChangeCount++
    playerStore.saveData()
    message.success(
      playerStore.nameChangeCount === 1 ? '道号修改成功！首次修改免费' : `道号修改成功！消耗${spiritStoneCost}颗灵石`
    )
    newName.value = ''
  }
</script>

<style scoped></style>