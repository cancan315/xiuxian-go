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
          <div class="donation-section">
            <div class="donation-text">
              本游戏免费，基于单机游戏《我的放置仙途》使用业余时间通过AI开发，耗时32天，欢迎打赏支持。
              <br /> 
              QQ玩家群：755301571
              <br />  开发计划：
              <br /> 
              1.灵根系统，5灵根、4灵根、3灵根、2灵根、天灵根、异灵根
              <br /> 
              2.功法系统，凡品功法、黄品功法、玄品功法、地品功法、天品功法
              <br /> 
              3.金丹系统，上品金丹、中品金丹、下品金丹、假丹
              <br /> 
              特别感谢鬼鬼赞助的云服务器
            </div>
            <div class="donation-images">
              <div class="donation-item">
                <div class="donation-title">微信赞助</div>
                <img 
                  src="https://xxxxj.s3.cn-north-1.jdcloud-oss.com/wx.jpg" 
                  alt="微信收款码"
                  class="donation-qrcode"
                  @click="showFullImage('wechat')"
                />
                <div class="donation-hint">点击图片可放大</div>
              </div>
              <div class="donation-item">
                <div class="donation-title">支付宝赞助</div>
                <img 
                  src="https://xxxxj.s3.cn-north-1.jdcloud-oss.com/zfb.jpg" 
                  alt="支付宝收款码"
                  class="donation-qrcode"
                  @click="showFullImage('alipay')"
                />
                <div class="donation-hint">点击图片可放大</div>
              </div>
            </div>
            <div class="donation-tips">
              <n-icon size="16" color="#f0a020">
                <AlertCircleOutline />
              </n-icon>
              扫码赞助时请备注您的游戏道号，感谢支持！
            </div>
          </div>
        </n-alert>
        <n-space>
          <n-button type="error" @click="qq = true">官方群聊</n-button>
          
        </n-space>
      </n-space>
    </n-card>
    
    <!-- 官方群聊弹窗 -->
    <n-modal preset="dialog" title="玩家交流群" v-model:show="qq">
      <n-card :bordered="false" size="huge" role="dialog" aria-modal="true">
        <n-space vertical>
          <n-input value="755301571" readonly type="text" />
        </n-space>
      </n-card>
    </n-modal>
    
    <!-- 收款码大图弹窗 -->
    <n-modal v-model:show="showQRCodeModal" :mask-closable="true">
      <n-card
        style="width: 600px; max-width: 90vw;"
        :bordered="false"
        size="huge"
        role="dialog"
        aria-modal="true"
        :title="currentQRCodeType === 'wechat' ? '微信收款码' : '支付宝收款码'"
      >
        <template #header-extra>
          <n-button quaternary circle @click="showQRCodeModal = false">
            <template #icon>
              <n-icon><Close /></n-icon>
            </template>
          </n-button>
        </template>
        <div class="full-qrcode-container">
          <img 
            :src="currentQRCodeType === 'wechat' 
              ? 'https://xxxxj.s3.cn-north-1.jdcloud-oss.com/wx.jpg' 
              : 'https://xxxxj.s3.cn-north-1.jdcloud-oss.com/zfb.jpg'" 
            :alt="currentQRCodeType === 'wechat' ? '微信收款码' : '支付宝收款码'"
            class="full-qrcode"
          />
        </div>
        <template #footer>
          <div class="qrcode-footer">
            <n-space justify="center">
              <n-button @click="currentQRCodeType = 'wechat'" :type="currentQRCodeType === 'wechat' ? 'primary' : 'default'">
                微信
              </n-button>
              <n-button @click="currentQRCodeType = 'alipay'" :type="currentQRCodeType === 'alipay' ? 'primary' : 'default'">
                支付宝
              </n-button>
            </n-space>
          </div>
        </template>
      </n-card>
    </n-modal>
  </div>
</template>

<script setup>
  // 修改为使用模块化store
  import { usePlayerInfoStore } from '../stores/playerInfo'
  import { useInventoryStore } from '../stores/inventory'
  import { useEquipmentStore } from '../stores/equipment'
  import { usePetsStore } from '../stores/pets'
  import { usePillsStore } from '../stores/pills'
  import { useSettingsStore } from '../stores/settings'
  import { useStatsStore } from '../stores/stats'
  import { ref } from 'vue'
  import { useDialog, useMessage } from 'naive-ui'
  import { AlertCircleOutline, Close } from '@vicons/ionicons5'

  const playerInfoStore = usePlayerInfoStore()
  const inventoryStore = useInventoryStore()
  const equipmentStore = useEquipmentStore()
  const petsStore = usePetsStore()
  const pillsStore = usePillsStore()
  const settingsStore = useSettingsStore()
  const statsStore = useStatsStore()
  
  const newName = ref('')
  const message = useMessage()
  const maxLength = 6 // 定义道号最大长度常量
  const dialog = useDialog()
  const version = __APP_VERSION__

  const qq = ref(false)
  const showQRCodeModal = ref(false)
  const currentQRCodeType = ref('wechat') // 'wechat' 或 'alipay'

  // 显示大图弹窗
  const showFullImage = (type) => {
    currentQRCodeType.value = type
    showQRCodeModal.value = true
  }

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
    const spiritStoneCost = playerInfoStore.nameChangeCount === 0 ? 0 : Math.pow(2, playerInfoStore.nameChangeCount) * 100
    // 第一次修改免费，之后需要消耗灵石
    if (playerInfoStore.nameChangeCount > 0) {
      if (inventoryStore.spiritStones < spiritStoneCost) {
        message.error(`灵石不足！修改道号需要${spiritStoneCost}颗灵石`)
        return
      }
      inventoryStore.spiritStones -= spiritStoneCost
    }
    playerInfoStore.renamePlayer(newName.value.trim())
    
    message.success(
      playerInfoStore.nameChangeCount === 1 ? '道号修改成功！首次修改免费' : `道号修改成功！消耗${spiritStoneCost}颗灵石`
    )
    newName.value = ''
  }
</script>

<style scoped>
  .settings-container {
    padding: 20px;
    max-width: 800px;
    margin: 0 auto;
  }

  .donation-section {
    margin-top: 8px;
  }

  .donation-text {
    margin-bottom: 16px;
    line-height: 1.5;
  }

  .donation-images {
    display: flex;
    justify-content: space-around;
    flex-wrap: wrap;
    gap: 24px;
    margin: 16px 0;
  }

  .donation-item {
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  .donation-title {
    font-weight: bold;
    margin-bottom: 8px;
    color: #f0a020;
  }

  .donation-qrcode {
    width: 150px;
    height: 150px;
    border-radius: 8px;
    border: 2px solid #f0a020;
    cursor: pointer;
    transition: transform 0.2s;
  }

  .donation-qrcode:hover {
    transform: scale(1.05);
  }

  .donation-hint {
    font-size: 12px;
    color: #666;
    margin-top: 4px;
  }

  .donation-tips {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-top: 12px;
    padding: 8px 12px;
    background-color: rgba(240, 160, 32, 0.1);
    border-radius: 4px;
    font-size: 14px;
    color: #f0a020;
  }

  .full-qrcode-container {
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 20px;
  }

  .full-qrcode {
    max-width: 100%;
    max-height: 400px;
    border-radius: 12px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  }

  .qrcode-footer {
    padding: 16px;
  }

  @media (max-width: 768px) {
    .donation-images {
      flex-direction: column;
      align-items: center;
    }
    
    .donation-qrcode {
      width: 120px;
      height: 120px;
    }
    
    .full-qrcode {
      max-height: 300px;
    }
  }
</style>