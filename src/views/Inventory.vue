<template>
  <n-layout>
    <n-layout-header bordered>
      <n-page-header>
        <template #title>背包</template>
      </n-page-header>
    </n-layout-header>
    <n-layout-content>
      <n-card :bordered="false">
        <n-tabs type="line">
          <n-tab-pane name="equipment" tab="装备">
            <n-grid :cols="2" :x-gap="12" :y-gap="8">
              <n-grid-item v-for="(type, index) in Object.keys(equipmentTypes)" :key="index">
                <n-card hoverable>
                  <template #header>
                    <n-space justify="space-between">
                      <span>{{ equipmentTypes[type] }}</span>
                      <n-button size="small" @click="() => showEquipmentList(type)">
                        更多
                      </n-button>
                    </n-space>
                  </template>
                  <p v-if="equipmentStore.equippedArtifacts[type]">
                    {{ equipmentStore.equippedArtifacts[type].name }}
                  </p>
                  <p v-else>未装备</p>
                  <template #footer>
                    <n-space justify="space-between">
                      <span>{{ equipmentTypes[type] }}</span>
                      <n-button
                        size="small"
                        type="info"
                        @click.stop="() => showEquippedEquipmentDetails(type)"
                        v-if="equipmentStore.equippedArtifacts[type]"
                      >
                        详细
                      </n-button>
                      <n-button
                        size="small"
                        type="error"
                        @click.stop="() => unequipItem(type)"
                        v-if="equipmentStore.equippedArtifacts[type]"
                      >
                        卸下
                      </n-button>
                    </n-space>
                  </template>
                </n-card>
              </n-grid-item>
            </n-grid>
          </n-tab-pane>
          <n-tab-pane name="herbs" tab="灵草">
            <n-grid :cols="2" :x-gap="12" :y-gap="8" v-if="groupedHerbs.length">
              <n-grid-item v-for="herb in groupedHerbs" :key="herb.id">
                <n-card hoverable>
                  <template #header>
                    <n-space justify="space-between">
                      <span>{{ herb.name }}({{ herb.count }})</span>
                    </n-space>
                  </template>
                  <p>{{ herb.description }}</p>
                </n-card>
              </n-grid-item>
            </n-grid>
            <n-empty v-else />
          </n-tab-pane>
          <n-tab-pane name="pills" tab="丹药">
            <n-grid :cols="2" :x-gap="12" :y-gap="8" v-if="groupedPills.length">
              <n-grid-item v-for="pill in groupedPills" :key="pill.id">
                <n-card hoverable>
                  <template #header>
                    <n-space justify="space-between">
                      <span>{{ pill.name }}({{ pill.count }})</span>
                      <n-button size="small" type="primary" @click="usePill(pill)">服用</n-button>
                    </n-space>
                  </template>
                  <p>{{ pill.description }}</p>
                </n-card>
              </n-grid-item>
            </n-grid>
            <n-empty v-else />
          </n-tab-pane>
          <n-tab-pane name="formulas" tab="丹方">
            <n-tabs type="segment">
              <n-tab-pane name="complete" tab="完整丹方">
                <n-grid :cols="2" :x-gap="12" :y-gap="8" v-if="groupedFormulas.complete.length">
                  <n-grid-item v-for="formula in groupedFormulas.complete" :key="formula.id">
                    <n-card hoverable>
                      <template #header>
                        <n-space justify="space-between">
                          <span>{{ formula.name }}</span>
                          <n-space>
                            <n-tag type="success" size="small">完整</n-tag>
                            <n-tag type="info" size="small">{{ pillGrades[formula.grade].name }}</n-tag>
                            <n-tag type="warning" size="small">{{ pillTypes[formula.type].name }}</n-tag>
                          </n-space>
                        </n-space>
                      </template>
                      <p>{{ formula.description }}</p>
                    </n-card>
                  </n-grid-item>
                </n-grid>
                <n-empty v-else />
              </n-tab-pane>
              <n-tab-pane name="incomplete" tab="残缺丹方">
                <n-grid :cols="2" :x-gap="12" :y-gap="8" v-if="groupedFormulas.incomplete.length">
                  <n-grid-item v-for="formula in groupedFormulas.incomplete" :key="formula.id">
                    <n-card hoverable>
                      <template #header>
                        <n-space justify="space-between">
                          <span>{{ formula.name }}</span>
                          <n-space>
                            <n-tag type="warning" size="small">残缺</n-tag>
                            <n-tag type="info" size="small">{{ pillGrades[formula.grade].name }}</n-tag>
                            <n-tag type="warning" size="small">{{ pillTypes[formula.type].name }}</n-tag>
                          </n-space>
                        </n-space>
                      </template>
                      <p>{{ formula.description }}</p>
                      <n-progress
                        type="line"
                        :percentage="Number(((formula.fragments / formula.fragmentsNeeded) * 100).toFixed(2))"
                        :show-indicator="true"
                        indicator-placement="inside"
                      >
                        收集进度: {{ formula.fragments }}/{{ formula.fragmentsNeeded }}
                      </n-progress>
                    </n-card>
                  </n-grid-item>
                </n-grid>
                <n-empty v-else />
              </n-tab-pane>
            </n-tabs>
          </n-tab-pane>
          <n-tab-pane name="pets" tab="灵宠">
            <n-space style="margin-bottom: 16px">
              <n-select
                v-model:value="selectedRarityToRelease"
                :options="options"
                placeholder="选择放生品阶"
                style="width: 150px"
              />
              <n-button
                @click="showBatchReleaseConfirm = true"
                :disabled="!petsStore.pets.length"
              >
                一键放生
              </n-button>
            </n-space>
            <n-modal v-model:show="showBatchReleaseConfirm" preset="dialog" title="批量放生确认" style="width: 600px">
              <p>
                确定要放生{{
                  selectedRarityToRelease === 'all' ? '所有' : petRarities[selectedRarityToRelease].name
                }}品阶的未出战灵宠吗？此操作不可撤销。
              </p>
              <n-space justify="end" style="margin-top: 16px">
                <n-button size="small" @click="showBatchReleaseConfirm = false">取消</n-button>
                <n-button size="small" type="error" @click="batchReleasePets">确认放生</n-button>
              </n-space>
            </n-modal>
            <n-pagination
              v-if="filteredPets.length > 12"
              v-model:page="currentPage"
              :page-size="pageSize"
              :item-count="filteredPets.length"
              @update:page-size="onPageSizeChange"
              :page-slot="7"
            />
            <div v-if="displayPets.length === 0" style="text-align: center; padding: 20px;">
              <n-empty description="暂无灵宠" />
              <p style="color: #999; margin-top: 10px;">通过抽奖可以获得灵宠</p>
            </div>
            <n-grid v-else-if="displayPets.length > 0" :cols="2" :x-gap="12" :y-gap="8" style="margin-top: 16px">
              <n-grid-item v-for="pet in displayPets" :key="pet.id">
                <n-card hoverable>
                  <template #header>
                    <n-space justify="space-between">
                      <span>{{ pet.name }}</span>
                      <n-button size="small" type="primary" @click="useItem(pet)">
                        {{ pet.isActive ? '召回' : '出战' }}
                      </n-button>
                    </n-space>
                  </template>
                  <p>{{ pet.description }}</p>
                  <n-space vertical>
                    <n-tag :style="{ color: pet.rarity && petRarities[pet.rarity] ? petRarities[pet.rarity].color : '#000000' }">
                      {{ pet.rarity && petRarities[pet.rarity] ? petRarities[pet.rarity].name : '未知品质' }}
                    </n-tag>
                    <n-space justify="space-between">
                      <n-text>等级: {{ pet.level || 1 }}</n-text>
                      <n-text>星级: {{ pet.star || 0 }}</n-text>
                      <n-button size="small" @click="showPetDetails(pet)">详情</n-button>
                    </n-space>
                  </n-space>
                </n-card>
              </n-grid-item>
            </n-grid>
          </n-tab-pane>
        </n-tabs>
      </n-card>
    </n-layout-content>
  </n-layout>
  <!-- 灵宠详情弹窗 -->
  <n-modal v-model:show="showPetModal" preset="dialog" title="灵宠详情" style="width: 600px">
    <template v-if="selectedPet">
      <n-descriptions bordered>
        <n-descriptions-item label="名称">{{ selectedPet.name }}</n-descriptions-item>
        <n-descriptions-item label="品质">
          <n-tag :style="{ color: selectedPet.rarity && petRarities[selectedPet.rarity] ? petRarities[selectedPet.rarity].color : '#000000' }">
            {{ selectedPet.rarity && petRarities[selectedPet.rarity] ? petRarities[selectedPet.rarity].name : '未知品质' }}
          </n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="等级">{{ selectedPet.level || 1 }}</n-descriptions-item>
        <n-descriptions-item label="星级">{{ selectedPet.star || 0 }}</n-descriptions-item>
        <n-descriptions-item label="境界">{{ Math.floor((selectedPet.star || 0) / 5) }}阶</n-descriptions-item>
      </n-descriptions>
      <n-divider>属性加成</n-divider>
      <n-descriptions bordered>
        <n-descriptions-item label="攻击加成">
          +{{ ((selectedPet.bonus?.attack || selectedPet.attackBonus || 0) * 100).toFixed(1) }}%
        </n-descriptions-item>
        <n-descriptions-item label="防御加成">
          +{{ ((selectedPet.bonus?.defense || selectedPet.defenseBonus || 0) * 100).toFixed(1) }}%
        </n-descriptions-item>
        <n-descriptions-item label="生命加成">
          +{{ ((selectedPet.bonus?.health || selectedPet.healthBonus || 0) * 100).toFixed(1) }}%
        </n-descriptions-item>
      </n-descriptions>
      <n-divider>灵宠属性</n-divider>
      <n-collapse>
        <n-collapse-item title="展开" name="1">
          <n-divider>基础属性</n-divider>
          <n-descriptions bordered :column="2">
            <n-descriptions-item label="攻击力">{{ selectedPet.combatAttributes?.attack || 0 }}</n-descriptions-item>
            <n-descriptions-item label="生命值">{{ selectedPet.combatAttributes?.health || 0 }}</n-descriptions-item>
            <n-descriptions-item label="防御力">{{ selectedPet.combatAttributes?.defense || 0 }}</n-descriptions-item>
            <n-descriptions-item label="速度">{{ selectedPet.combatAttributes?.speed || 0 }}</n-descriptions-item>
          </n-descriptions>
          <n-divider>战斗属性</n-divider>
          <n-descriptions bordered :column="3">
            <n-descriptions-item label="暴击率">
              {{ ((selectedPet.combatAttributes?.critRate || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="连击率">
              {{ ((selectedPet.combatAttributes?.comboRate || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="反击率">
              {{ ((selectedPet.combatAttributes?.counterRate || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="眩晕率">
              {{ ((selectedPet.combatAttributes?.stunRate || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="闪避率">
              {{ ((selectedPet.combatAttributes?.dodgeRate || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="吸血率">
              {{ ((selectedPet.combatAttributes?.vampireRate || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
          </n-descriptions>
          <n-divider>战斗抗性</n-divider>
          <n-descriptions bordered :column="3">
            <n-descriptions-item label="抗暴击">
              {{ ((selectedPet.combatAttributes?.critResist || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="抗连击">
              {{ ((selectedPet.combatAttributes?.comboResist || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="抗反击">
              {{ ((selectedPet.combatAttributes?.counterResist || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="抗眩晕">
              {{ ((selectedPet.combatAttributes?.stunResist || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="抗闪避">
              {{ ((selectedPet.combatAttributes?.dodgeResist || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="抗吸血">
              {{ ((selectedPet.combatAttributes?.vampireResist || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
          </n-descriptions>
          <n-divider>特殊属性</n-divider>
          <n-descriptions bordered :column="3">
            <n-descriptions-item label="强化治疗">
              {{ ((selectedPet.combatAttributes?.healBoost || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="强化爆伤">
              {{ ((selectedPet.combatAttributes?.critDamageBoost || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="弱化爆伤">
              {{ ((selectedPet.combatAttributes?.critDamageReduce || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="最终增伤">
              {{ ((selectedPet.combatAttributes?.finalDamageBoost || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="最终减伤">
              {{ ((selectedPet.combatAttributes?.finalDamageReduce || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="战斗属性提升">
              {{ ((selectedPet.combatAttributes?.combatBoost || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
            <n-descriptions-item label="战斗抗性提升">
              {{ ((selectedPet.combatAttributes?.resistanceBoost || 0) * 100).toFixed(1) }}%
            </n-descriptions-item>
          </n-descriptions>
        </n-collapse-item>
      </n-collapse>
      <n-divider>操作</n-divider>
      <n-space vertical>
        <n-space justify="space-between">
          <span>升级（消耗{{ getUpgradeCost(selectedPet) }} / {{ playerInfoStore.petEssence }}灵宠精华）</span>
          <n-button size="small" type="primary" @click="upgradePet(selectedPet)" :disabled="!canUpgrade(selectedPet)">
            升级
          </n-button>
        </n-space>
        <n-space justify="space-between">
          <span>升星（同品质灵宠，相同名字成功率100%，不同名字成功率30%）</span>
          <n-select
            v-model:value="selectedFoodPet"
            :options="getAvailableFoodPets(selectedPet)"
            placeholder="选择升星材料"
            style="width: 200px"
          />
          <n-button size="small" type="warning" @click="evolvePet(selectedPet)" :disabled="!selectedFoodPet">
            升星
          </n-button>
        </n-space>
        <n-space justify="space-between">
          <span>放生灵宠（不会返还已消耗的道具）</span>
          <n-button size="small" type="error" @click="confirmReleasePet(selectedPet)">放生灵宠</n-button>
          <n-modal v-model:show="showReleaseConfirm" preset="dialog" title="灵宠放生" style="width: 600px">
            <template v-if="petToRelease">
              <p>确定要放生 {{ petToRelease.name }} 吗？此操作不可撤销，且不会返还已消耗的道具。</p>
              <n-space justify="end" style="margin-top: 16px">
                <n-button size="small" @click="cancelReleasePet">取消</n-button>
                <n-button size="small" type="error" @click="releasePet">确认放生</n-button>
              </n-space>
            </template>
          </n-modal>
        </n-space>
      </n-space>
    </template>
  </n-modal>
  <!-- 装备列表弹窗 -->
  <n-modal
    v-model:show="showEquipmentModal"
    preset="dialog"
    :title="`${equipmentTypes[selectedEquipmentType]}列表`"
    style="width: 800px"
    @after-leave="onCloseEquipmentModal"
  >
    <n-space vertical>
      <n-space justify="space-between">
        <n-select v-model:value="selectedQuality" :options="qualityOptions" style="width: 150px" />
        <n-button type="warning" :disabled="equipmentList.length === 0" @click="batchSellEquipments">一键卖出</n-button>
      </n-space>
      <n-pagination
        v-model:page="currentEquipmentPage"
        :page-size="equipmentPageSize"
        :item-count="filteredEquipmentList.length"
        v-if="equipmentList.length > 8"
        @update:page-size="onEquipmentPageSizeChange"
        :page-slot="7"
      />
      <n-grid :cols="2" :x-gap="12" :y-gap="8" v-if="pagedEquipmentList.length">
        <n-grid-item v-for="equipment in pagedEquipmentList" :key="equipment.id" @click="showEquipmentDetails(equipment)">
          <n-card hoverable>
            <template #header>
              <n-space justify="space-between">
                <span @click.stop="() => logClickEvent('装备名称', equipment)">{{ equipment.name }}</span>
                <n-button size="small" type="warning" @click.stop="sellEquipment(equipment)">卖出</n-button>
              </n-space>
            </template>
            <n-space vertical>
              <n-tag 
                :style="{ color: (equipment.quality && equipmentQualities[equipment.quality]?.color) || '#000000' }"
                @click.stop="() => logClickEvent('装备品质标签', equipment)"
              >
                {{ (equipment.quality && equipmentQualities[equipment.quality]?.name) || '未知品质' }}
              </n-tag>
              <n-text @click.stop="() => logClickEvent('境界要求文本', equipment)">境界要求：{{ getRealmPeriodName(equipment.requiredRealm) || '未知境界' }}</n-text>
              <!-- 显示装备状态 -->
              <n-tag v-if="equipment.equipped" type="success" @click.stop="() => logClickEvent('装备状态标签', equipment)">
                已装备
              </n-tag>
            </n-space>
          </n-card>
        </n-grid-item>
      </n-grid>
      <n-empty description="没有任何装备" v-else></n-empty>
    </n-space>
  </n-modal>
  <!-- 装备详情弹窗 -->
  <n-modal v-model:show="showEquipmentDetailModal" preset="dialog" :title="selectedEquipment?.name || '装备详情'">
    <n-descriptions bordered>
      <n-descriptions-item label="品质">
        <span :style="{ color: (selectedEquipment?.quality && equipmentQualities[selectedEquipment.quality]?.color) || '#000000' }">
          {{ (selectedEquipment?.quality && equipmentQualities[selectedEquipment.quality]?.name) || '未知品质' }}
        </span>
      </n-descriptions-item>
      <n-descriptions-item label="类型">
        {{ equipmentTypes[selectedEquipment?.type] }}
      </n-descriptions-item>
      <n-descriptions-item label="强化等级">+{{ selectedEquipment?.enhanceLevel || 0 }}</n-descriptions-item>
      <template v-if="selectedEquipment?.stats">
        <n-descriptions-item v-for="(value, stat) in selectedEquipment.stats" :key="stat" :label="getStatName(stat)">
          {{ formatStatValue(stat, value) }}
        </n-descriptions-item>
      </template>
    </n-descriptions>
    <div
      class="stats-comparison"
      v-if="equipmentComparison && selectedEquipment && selectedEquipment.id != equipmentStore.equippedArtifacts[selectedEquipment.type]?.id"
    >
      <n-divider>属性对比</n-divider>
      <n-table :bordered="false" :single-line="false">
        <thead>
          <tr>
            <th>属性</th>
            <th>当前装备</th>
            <th>选中装备</th>
            <th>属性变化</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(comparison, stat) in equipmentComparison" :key="stat">
            <td>{{ getStatName(stat) }}</td>
            <td>{{ formatStatValue(stat, comparison.current) }}</td>
            <td>{{ formatStatValue(stat, comparison.selected) }}</td>
            <td>
              <n-gradient-text :type="comparison.isPositive ? 'success' : 'error'">
                {{ comparison.isPositive ? '+' : '' }}{{ formatStatValue(stat, comparison.diff) }}
              </n-gradient-text>
            </td>
          </tr>
        </tbody>
      </n-table>
    </div>
    <template #action>
      <n-space justify="space-between">
        <n-space>
          <n-button
            type="primary"
            @click="showEnhanceConfirm = true"
            :disabled="(selectedEquipment?.enhanceLevel || 0) >= 100"
          >
            强化
          </n-button>
          <n-button type="info" :disabled="playerInfoStore.refinementStones === 0" @click="handleReforgeEquipment">
            洗练
          </n-button>
        </n-space>
        <n-space>
          <n-button
            @click="equipItem(selectedEquipment)"
            :disabled="playerInfoStore.level < selectedEquipment?.requiredRealm"
            v-if="selectedEquipment && selectedEquipment.id != equipmentStore.equippedArtifacts[selectedEquipment.type]?.id"
          >
            装备
          </n-button>
          <n-button
            @click="unequipItem(selectedEquipment?.type)"
            :disabled="playerInfoStore.level < selectedEquipment?.requiredRealm"
            v-else-if="selectedEquipment"
          >
            卸下
          </n-button>
          <n-button
            type="error"
            @click="sellEquipment(selectedEquipment)"
            v-if="selectedEquipment && selectedEquipment.id != equipmentStore.equippedArtifacts[selectedEquipment.type]?.id"
          >
            出售
          </n-button>
        </n-space>
      </n-space>
    </template>
  </n-modal>
  <!-- 强化确认弹窗 -->
  <n-modal v-model:show="showEnhanceConfirm" preset="dialog" title="装备强化">
    <n-space vertical>
      <p>是否消耗 {{ ((selectedEquipment?.enhanceLevel || 0) + 1) * 10 }} 强化石强化装备？</p>
      <p>当前强化石数量：{{ inventoryStore.reinforceStones }}</p>
    </n-space>
    <template #action>
      <n-space justify="end">
        <n-button @click="showEnhanceConfirm = false">取消</n-button>
        <n-button
          type="primary"
          @click="handleEnhanceEquipment"
          :disabled="inventoryStore.reinforceStones < ((selectedEquipment?.enhanceLevel || 0) + 1) * 10"
        >
          确认强化
        </n-button>
      </n-space>
    </template>
  </n-modal>
  <!-- 洗练确认弹窗 -->
  <n-modal v-model:show="showReforgeConfirm" preset="dialog" title="洗练结果确认">
    <template v-if="reforgeResult">
      <div class="reforge-compare">
        <div class="old-stats">
          <h3>原始属性</h3>
          <div v-for="(value, key) in reforgeResult.oldStats" :key="key">
            {{ getStatName(key) }}: {{ formatStatValue(key, value) }}
          </div>
        </div>
        <div class="new-stats">
          <h3>新属性</h3>
          <div v-for="(value, key) in reforgeResult.newStats" :key="key">
            {{ getStatName(key) }}: {{ formatStatValue(key, value) }}
          </div>
        </div>
      </div>
    </template>
    <template #action>
      <n-button type="primary" @click="confirmReforgeResult(true)">确认新属性</n-button>
      <n-button @click="confirmReforgeResult(false)">保留原属性</n-button>
    </template>
  </n-modal>
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
  import { usePersistenceStore } from '../stores/persistence'
  import { ref, computed, onMounted } from 'vue'
  import { useMessage } from 'naive-ui'
  import { getStatName, formatStatValue } from '../plugins/stats'
  import { getRealmName, getRealmPeriodName } from '../plugins/realm'
  import { pillRecipes, pillGrades, pillTypes, calculatePillEffect } from '../plugins/pills'
  import { itemQualities } from '../plugins/itemQualities'
  import APIService from '../services/api.js'
  
  const playerInfoStore = usePlayerInfoStore()
  const inventoryStore = useInventoryStore()
  const equipmentStore = useEquipmentStore()
  const petsStore = usePetsStore()
  const pillsStore = usePillsStore()
  const settingsStore = useSettingsStore()
  const statsStore = useStatsStore()
  const persistenceStore = usePersistenceStore()
  
  const message = useMessage()
  
  // 日志记录函数
  const logClickEvent = (elementType, equipment) => {
    console.log(`[Inventory] 用户点击了装备列表中的${elementType}: ${equipment.name} (${equipment.id})`)
  }
  
  // 装备类型
  const equipmentTypes = {
    faqi: '法宝',
    guanjin: '冠巾',
    daopao: '道袍',
    yunlv: '云履',
    fabao: '本命法宝'
  }

  // 装备类型到equipType的映射
  const equipmentTypeMap = {
    faqi: '法宝',
    guanjin: '冠巾',
    daopao: '道袍',
    yunlv: '云履',
    fabao: '法宝'
  }
  
  // 装备品质配置（使用统一配置）
  const equipmentQualities = itemQualities.equipment

  // 分页相关
  const currentPage = ref(1)
  const pageSize = ref(12)
  
  // 装备弹窗相关
  const showEquipmentModal = ref(false)
  const selectedEquipmentType = ref(null)
  
  // 本地装备列表（从后端直接获取）
  const localEquipmentList = ref([])
  
  // 装备列表（只使用本地列表）
  const equipmentList = computed(() => {
    // 只使用本地装备列表数据
    console.log(`[Inventory] 使用本地装备列表，数量: ${localEquipmentList.value.length}`)
    return localEquipmentList.value
  })
  
  // 过滤后的灵宠列表
  const filteredPets = computed(() => {
    console.log('[Inventory] 计算过滤后的灵宠列表，当前灵宠总数:', petsStore.pets.length);
    console.log('[Inventory] 当前灵宠列表:', petsStore.pets);
    const pets = petsStore.pets
    if (selectedRarityToRelease.value === 'all') {
      console.log('[Inventory] 不过滤品质，返回所有灵宠');
      return pets
    }
    console.log('[Inventory] 按品质过滤灵宠:', selectedRarityToRelease.value);
    const filtered = pets.filter(pet => pet.rarity === selectedRarityToRelease.value);
    console.log('[Inventory] 过滤后的灵宠数量:', filtered.length);
    return filtered;
  })

  // 当前页显示的灵宠
  const displayPets = computed(() => {
    const start = (currentPage.value - 1) * pageSize.value
    const end = start + pageSize.value
    console.log('[Inventory] 计算当前页显示的灵宠，范围:', start, '-', end);
    let displayed = filteredPets.value.slice(start, end);
    
    // 对灵宠列表进行排序，已出战的灵宠排在第一位
    displayed = displayed.sort((a, b) => {
      // 已出战的灵宠排在前面
      if (a.isActive && !b.isActive) return -1;
      if (!a.isActive && b.isActive) return 1;
      return 0;
    });
    
    console.log('[Inventory] 当前页显示的灵宠数量:', displayed.length);
    
    // 打印排序后的灵宠列表用于调试
    console.log('[Inventory] 当前页灵宠排序结果:');
    displayed.forEach((pet, index) => {
      console.log(`  ${index + 1}. ${pet.name}(${pet.id}): isActive=${pet.isActive}`);
    });
    
    return displayed;
  })

  // 页大小改变处理
  const onPageSizeChange = size => {
    pageSize.value = size
    currentPage.value = 1
  }

  // 选中的放生品阶
  const selectedRarityToRelease = ref('all')

  // 使用丹药
  const usePill = pill => {
    const result = inventoryStore.useItem(pill, pillsStore, petsStore, playerInfoStore, persistenceStore)
    if (result.success) {
      message.success(result.message)
    } else {
      message.error(result.message)
    }
  }

  // 灵宠品质配置（使用统一配置）
  const petRarities = itemQualities.pet

  // 灵宠详情相关
  const showPetModal = ref(false)
  const selectedPet = ref(null)
  const selectedFoodPet = ref(null)

  // 放生确认弹窗
  const showReleaseConfirm = ref(false)
  const showBatchReleaseConfirm = ref(false)
  const petToRelease = ref(null)

  // 显示放生确认弹窗
  const confirmReleasePet = pet => {
    petToRelease.value = pet
    showReleaseConfirm.value = true
  }

  // 取消放生
  const cancelReleasePet = () => {
    petToRelease.value = null
    showReleaseConfirm.value = false
  }

  // 执行放生
  const releasePet = async () => {
    if (petToRelease.value) {
      const token = getAuthToken()
      try {
        // 调用API放生灵宠
        const response = await APIService.deletePets(token, [petToRelease.value.id])
        if (response.success) {
          message.success('已放生灵宠')
        } else {
          message.error(response.message || '放生失败')
        }
      } catch (error) {
        console.error('放生灵宠失败:', error)
        message.error('放生灵宠失败: ' + error.message)
      }
      
      // 关闭所有相关弹窗
      showReleaseConfirm.value = false
      showPetModal.value = false
      petToRelease.value = null
    }
  }

  // 显示灵宠详情
  const showPetDetails = pet => {
    selectedPet.value = pet
    selectedFoodPet.value = null
    showPetModal.value = true
  }

  // 计算灵宠属性加成
  const getPetBonus = (pet) => {
    return { attack: 0, defense: 0, health: 0 }
  }

  // 获取升级所需精华数量
  const getUpgradeCost = pet => {
    return (pet.level || 1) * 10
  }

  // 检查是否可以升级
  const canUpgrade = pet => {
    const cost = getUpgradeCost(pet)
    return playerInfoStore.petEssence >= cost
  }

  // 获取可用作升星材料的灵宠列表
  const getAvailableFoodPets = pet => {
    if (!pet) return []
    return petsStore.pets
      .filter(
        item =>
          item.id !== pet.id &&
          item.star === pet.star &&
          item.rarity === pet.rarity
      )
      .map(item => ({
        label: `${item.name} (${item.level || 1}级 ${item.star || 0}星)${item.name !== pet.name ? ' [成功率30%]' : ''}`,
        value: item.id
      }))
  }

  // 升级灵宠
  const upgradePet = async pet => {
    const token = getAuthToken()
    try {
      const response = await APIService.upgradePet(token, pet.id, getUpgradeCost(pet))
      if (response.success) {
        message.success('升级成功')
      } else {
        message.error(response.message || '升级失败')
      }
    } catch (error) {
      console.error('升级灵宠失败:', error)
      message.error('升级灵宠失败: ' + error.message)
    }
  }

  // 升星灵宠
  const evolvePet = async pet => {
    if (!selectedFoodPet.value) {
      message.error('请选择用于升星的灵宠')
      return
    }
    // 通过id查找对应的灵宠对象
    const foodPet = petsStore.pets.find(item => item.id === selectedFoodPet.value)
    if (!foodPet) {
      message.error('升星材料灵宠不存在')
      return
    }
    
    const token = getAuthToken()
    try {
      const response = await APIService.evolvePet(token, pet.id, foodPet.id)
      if (response.success) {
        message.success('升星成功')
        selectedFoodPet.value = null
        showPetModal.value = false
      } else {
        message.error(response.message || '升星失败')
      }
    } catch (error) {
      console.error('升星灵宠失败:', error)
      message.error('升星灵宠失败: ' + error.message)
    }
  }

  // 卸下装备
  const unequipItem = async slot => {
    // 添加日志记录
    const currentEquipment = equipmentStore.equippedArtifacts[slot]
    if (currentEquipment) {
      console.log(`[Inventory] 玩家尝试卸下装备: ${currentEquipment.name} (${currentEquipment.id}) 从 ${equipmentTypes[slot]} 槽位`)
      console.log('[Inventory] 卸下前的装备信息:', JSON.stringify(currentEquipment, null, 2))
    }
    
    const token = getAuthToken()
    // 使用明确的参数调用 unequipArtifact 方法
    console.log(`[Inventory] 调用卸下装备接口，槽位: ${slot}`);
    const result = await equipmentStore.unequipArtifact(
      slot,                   // 装备槽位
      inventoryStore,         // 背包存储
      persistenceStore,       // 持久化存储
      token                   // 认证令牌
    );
    console.log('[Inventory] 卸下装备接口返回结果:', JSON.stringify(result, null, 2));
    
    if (result.success) {
      console.log('[Inventory] 卸下装备成功，开始处理结果');
      showEquipmentDetailModal.value = false
      message.success('当前装备已卸下')
      // 卸下装备后刷新已装备列表和背包列表
      await loadEquippedArtifacts()
      // 清空本地装备列表以触发重新加载
      localEquipmentList.value = []
      
      // 更新玩家属性
      console.log('[Inventory] 检查返回结果中是否包含用户属性:', !!result.user);
      if (result.user) {
        console.log('[Inventory] 卸下装备前玩家属性:', JSON.stringify({
          baseAttributes: playerInfoStore.baseAttributes,
          combatAttributes: playerInfoStore.combatAttributes,
          combatResistance: playerInfoStore.combatResistance,
          specialAttributes: playerInfoStore.specialAttributes
        }, null, 2));
        
        playerInfoStore.$patch({
          baseAttributes: result.user.baseAttributes,
          combatAttributes: result.user.combatAttributes,
          combatResistance: result.user.combatResistance,
          specialAttributes: result.user.specialAttributes
        });
        
        console.log('[Inventory] 卸下装备后玩家属性:', JSON.stringify({
          baseAttributes: playerInfoStore.baseAttributes,
          combatAttributes: playerInfoStore.combatAttributes,
          combatResistance: playerInfoStore.combatResistance,
          specialAttributes: playerInfoStore.specialAttributes
        }, null, 2));
      } else {
        console.warn('[Inventory] 卸下装备返回结果中未包含用户属性');
      }
      
      // 添加成功日志记录
      if (result.equipment) {
        console.log(`[Inventory] 成功卸下装备: ${result.equipment.name} (${result.equipment.id}) 从 ${equipmentTypes[slot]} 槽位`)
        console.log('[Inventory] 卸下后的装备信息:', JSON.stringify(result.equipment, null, 2))
        // 更新选中的装备信息
        if (selectedEquipment.value && selectedEquipment.value.id === result.equipment.id) {
          selectedEquipment.value = result.equipment
        }
      }
    } else {
      message.error(result.message || '卸下装备失败')
      // 添加错误日志记录
      if (currentEquipment) {
        console.error(`[Inventory] 卸下装备失败: ${currentEquipment.name} (${currentEquipment.id}) 从 ${equipmentTypes[slot]} 槽位`)
      }
      // 即使卸下失败也要清除缓存以确保列表刷新
      localEquipmentList.value = []
    }
  }

  // 使用装备
  const equipItem = async equipment => {
    // 添加日志记录
    console.log(`[Inventory] 玩家尝试装备物品: ${equipment.name} (${equipment.id})`)
    
    const token = getAuthToken()
    console.log(`[Inventory] 获取到的认证令牌:`, { 
      tokenAvailable: !!token, 
      tokenLength: token ? token.length : 0 
    })
    
    // 添加令牌有效性检查
    if (!token) {
      message.error('认证已过期，请重新登录')
      console.error('[Inventory] 装备穿戴失败：认证令牌缺失')
      return
    }
    
    // 使用 equipment.equipType 作为装备槽位参数，确保传递正确的装备槽位
    const result = await equipmentStore.equipArtifact(
      equipment,              // 装备对象
      equipment.equipType,    // 装备槽位
      inventoryStore,         // 背包存储
      persistenceStore,       // 持久化存储
      playerInfoStore,        // 玩家信息存储
      token                   // 认证令牌
    )
    
    if (result.success) {
      message.success(result.message)
      showEquipmentModal.value = false
      showEquipmentDetailModal.value = false
      // 装备成功后刷新已装备列表
      await loadEquippedArtifacts()
      // 清空本地装备列表以触发重新加载
      localEquipmentList.value = []
      
      // 更新玩家属性
      if (result.user) {
        console.log('[Inventory] 装备穿戴前玩家属性:', JSON.stringify({
          baseAttributes: playerInfoStore.baseAttributes,
          combatAttributes: playerInfoStore.combatAttributes,
          combatResistance: playerInfoStore.combatResistance,
          specialAttributes: playerInfoStore.specialAttributes
        }, null, 2));
        
        playerInfoStore.$patch({
          baseAttributes: result.user.baseAttributes,
          combatAttributes: result.user.combatAttributes,
          combatResistance: result.user.combatResistance,
          specialAttributes: result.user.specialAttributes
        });
        
        console.log('[Inventory] 装备穿戴后玩家属性:', JSON.stringify({
          baseAttributes: playerInfoStore.baseAttributes,
          combatAttributes: playerInfoStore.combatAttributes,
          combatResistance: playerInfoStore.combatResistance,
          specialAttributes: playerInfoStore.specialAttributes
        }, null, 2));
      }
      
      // 添加成功日志记录
      if (result.equipment) {
        console.log(`[Inventory] 成功装备物品: ${result.equipment.name} (${result.equipment.id})`)
        console.log('[Inventory] 装备成功后的装备信息:', JSON.stringify(result.equipment, null, 2))
        // 更新选中的装备信息
        if (selectedEquipment.value && selectedEquipment.value.id === result.equipment.id) {
          selectedEquipment.value = result.equipment
        }
      }
    } else {
      message.error(result.message || '装备失败')
      // 添加错误日志记录
      console.error(`[Inventory] 装备物品失败: ${equipment.name} (${equipment.id}), 原因: ${result.message || '未知错误'}`)
      console.error(`[Inventory] 尝试装备的物品详情:`, equipment)
      console.error(`[Inventory] 服务器返回的错误详情:`, result)
      // 即使装备失败也要清除缓存以确保列表刷新
      localEquipmentList.value = []
    }
  }

  // 导入必要的模块
  import { getAuthToken } from '../stores/db'

  // 修改装备列表显示方法
  const showEquipmentList = async (type) => {
    selectedEquipmentType.value = type
    // 添加日志记录
    console.log(`[Inventory] 玩家查看装备列表: ${equipmentTypes[type]}`)
    
    // 每次点击"更多"按钮都从后端获取最新数据
    try {
      const token = getAuthToken() // 使用正确的获取认证令牌方法
      if (!token) {
        message.error('未找到认证令牌')
        console.error('[Inventory] 未找到认证令牌，无法获取装备列表')
        return
      }
      
      console.log(`[Inventory] 开始获取装备列表，类型: ${type}`)
      // 使用新的API服务方法获取装备数据，获取所有装备（已装备和未装备）
      // 修改为使用equipType字段过滤
      const data = await APIService.getPlayerInventory(token, { equipType: type })
      
      // 直接使用从后端获取的装备列表，而不是更新 inventoryStore.items
      // 保存装备列表到本地变量
      // 注意：后端已经根据类型过滤了数据，不需要再在前端过滤
      localEquipmentList.value = data.items;
      console.log(`[Inventory] 成功将装备数据存储到 localEquipmentList，数量: ${data.items.length}，类型: ${type}`)
    } catch (error) {
      console.error('获取装备数据时发生错误:', error)
      message.error('获取装备数据时发生错误')
      // 添加错误日志记录
      console.error(`[Inventory] 获取装备数据时发生错误: ${equipmentTypes[type]}, 错误: ${error.message}`)
    }
    
    showEquipmentModal.value = true
    
    // 添加日志记录
    console.log(`[Inventory] 显示装备列表: ${equipmentTypes[type]}, 数量: ${localEquipmentList.value.length}`)
  }
  
  // 获取已装备的装备列表
  const loadEquippedArtifacts = async () => {
    try {
      const token = getAuthToken()
      if (!token) {
        console.error('[Inventory] 未找到认证令牌，无法获取已装备列表')
        return
      }
      
      console.log('[Inventory] 开始获取已装备的装备列表')
      const data = await APIService.getPlayerInventory(token, { equipped: 'true' })
      
      // 更新已装备的装备状态
      data.items.forEach(item => {
        if (item.slot) {
          equipmentStore.equippedArtifacts[item.slot] = item
        }
      })
      
      // 清空没有装备的槽位
      Object.keys(equipmentStore.equippedArtifacts).forEach(slot => {
        const isEquipped = data.items.some(item => item.slot === slot);
        if (!isEquipped) {
          equipmentStore.equippedArtifacts[slot] = null;
        }
      });
      
      console.log(`[Inventory] 成功获取已装备的装备列表，数量: ${data.items.length}`)
    } catch (error) {
      console.error('[Inventory] 获取已装备装备列表时发生错误:', error)
    }
  }
  
  // 页面加载时获取已装备的装备
  onMounted(() => {
    console.log('[Inventory] 页面挂载，开始加载已装备的装备和灵宠列表');
    loadEquippedArtifacts();
    
    // 加载灵宠列表
    loadPetList();
  })
  
  // 加载灵宠列表
  const loadPetList = async () => {
    try {
      console.log('[Inventory] 开始加载灵宠列表');
      const token = getAuthToken();
      if (!token) {
        console.error('[Inventory] 未找到认证令牌，无法获取灵宠列表');
        return;
      }
      
      const response = await APIService.getPlayerData(token);
      console.log('[Inventory] 获取到玩家完整数据:', response);
      
      if (response.pets) {
        console.log('[Inventory] 更新灵宠列表，数量:', response.pets.length);
        petsStore.pets = response.pets;
        
        // 打印灵宠列表状态用于调试
        console.log('[Inventory] 当前灵宠列表状态:');
        petsStore.pets.forEach(pet => {
          console.log(`  ${pet.name}(${pet.id}): isActive=${pet.isActive}`);
        });
      } else {
        console.warn('[Inventory] 响应中未包含灵宠数据');
      }
    } catch (error) {
      console.error('[Inventory] 加载灵宠列表时发生错误:', error);
    }
  }
  
  // 查看已装备的装备详情
  const showEquippedEquipmentDetails = async (type) => {
    const equippedItem = equipmentStore.equippedArtifacts[type];
    if (equippedItem) {
      // 直接显示已装备的装备详情
      selectedEquipment.value = equippedItem;
      showEquipmentDetailModal.value = true;
      console.log(`[Inventory] 显示已装备的装备详情: ${equippedItem.name} (${equippedItem.id})`);
    } else {
      message.info('该位置暂未装备任何装备');
      console.log(`[Inventory] 尝试查看未装备位置的详情: ${type}`);
    }
  }
  
  // 关闭装备列表弹窗时，清空本地装备列表
  const onCloseEquipmentModal = () => {
    console.log(`[Inventory] 关闭装备列表弹窗，清空本地装备列表，原数量: ${localEquipmentList.value.length}`)
    localEquipmentList.value = []
  }

  // 装备品质选项
  const qualityOptions = computed(() => {
    // 直接查询后端API获取装备列表，而不是使用 inventoryStore.items
    const equipmentsByQuality = {}
    equipmentList.value
      .forEach(item => {
        equipmentsByQuality[item.quality] = (equipmentsByQuality[item.quality] || 0) + 1
      })
    
    // 使用统一的装备品质配置
    const equipmentQualities = itemQualities.equipment
    return [
      { label: '全部品质', value: 'all' },
      { label: equipmentQualities.mythic.name, value: 'mythic', disabled: !equipmentsByQuality['mythic'] },
      { label: equipmentQualities.legendary.name, value: 'legendary', disabled: !equipmentsByQuality['legendary'] },
      { label: equipmentQualities.epic.name, value: 'epic', disabled: !equipmentsByQuality['epic'] },
      { label: equipmentQualities.rare.name, value: 'rare', disabled: !equipmentsByQuality['rare'] },
      { label: equipmentQualities.uncommon.name, value: 'uncommon', disabled: !equipmentsByQuality['uncommon'] },
      { label: equipmentQualities.common.name, value: 'common', disabled: !equipmentsByQuality['common'] }
    ]
  })
  
  // 过滤后的装备列表
  const filteredEquipmentList = computed(() => {
    console.log('[Inventory] 计算过滤后的装备列表')
    // 直接使用从后端获取的装备列表，而不是 inventoryStore.items
    let list = equipmentList.value.filter(item => {
      // 不再需要检查类型，因为后端已经做了类型过滤
      // 只需要检查品质
      if (selectedQuality.value !== 'all' && item.quality !== selectedQuality.value) return false
      return true
    })
    
    // 对列表进行排序，已装备的装备排在前面
    list.sort((a, b) => {
      // 已装备的装备排在前面
      if (a.equipped && !b.equipped) return -1
      if (!a.equipped && b.equipped) return 1
      return 0
    })
    
    console.log(`[Inventory] 过滤后的装备列表数量: ${list.length}`)
    return list
  })

  // 批量卖出装备
  const batchSellEquipments = async () => {
    const token = getAuthToken()
    const result = await equipmentStore.batchSellEquipments(
      selectedQuality.value === 'all' ? null : selectedQuality.value,
      selectedEquipmentType.value,
      inventoryStore,
      persistenceStore,
      token
    )
    
    if (result.success) {
      message.success(result.message)
      // 批量卖出成功后刷新本地装备列表缓存
      localEquipmentList.value = []
      // 重新从后端获取装备列表
      await showEquipmentList(selectedEquipmentType.value)
      console.log(`[Inventory] 成功批量卖出装备, 品质: ${selectedQuality.value === 'all' ? '全部' : selectedQuality.value}, 类型: ${selectedEquipmentType.value}`)
    } else {
      message.error(result.message || '批量卖出失败')
      console.error(`[Inventory] 批量卖出装备失败, 品质: ${selectedQuality.value === 'all' ? '全部' : selectedQuality.value}, 类型: ${selectedEquipmentType.value}, 原因: ${result.message || '未知错误'}`)
    }
  }

  // 卖出单件装备
  const sellEquipment = async equipment => {
    // 添加日志记录
    console.log(`[Inventory] 玩家尝试卖出装备: ${equipment.name} (${equipment.id})`)
    
    const token = getAuthToken()
    const result = await equipmentStore.sellEquipment(equipment, inventoryStore, persistenceStore, token)
    if (result.success) {
      message.success(result.message)
      showEquipmentDetailModal.value = false
      // 添加成功日志记录
      console.log(`[Inventory] 成功卖出装备: ${equipment.name} (${equipment.id})`)
      // 刷新本地装备列表缓存
      localEquipmentList.value = []
      // 重新从后端获取装备列表
      await showEquipmentList(selectedEquipmentType.value)
    } else {
      message.error(result.message || '卖出失败')
      // 添加错误日志记录
      console.error(`[Inventory] 卖出装备失败: ${equipment.name} (${equipment.id}), 原因: ${result.message || '未知错误'}`)
    }
  }

  // 修改装备详情显示方法
  const showEquipmentDetails = async (equipment) => {
    // 添加日志记录
    if (equipment && equipment.id) {
      console.log(`[Inventory] 玩家查看装备详情: ${equipment.name} (${equipment.id})`)
    }
    
    // 总是通过后端API获取装备详情，确保获取最新状态
    try {
      const token = getAuthToken() // 使用正确的获取认证令牌方法
      if (!token) {
        message.error('未找到认证令牌')
        console.error('[Inventory] 未找到认证令牌，无法获取装备详情')
        return
      }
      
      const itemId = equipment?.id || selectedEquipment.value?.id
      if (!itemId) {
        message.error('无效的装备ID')
        console.error('[Inventory] 无效的装备ID，无法获取装备详情')
        return
      }
      
      console.log(`[Inventory] 开始获取装备详情，ID: ${itemId}`)
      const response = await APIService.getItemDetails(token, itemId)
      selectedEquipment.value = response.item
      showEquipmentDetailModal.value = true
      
      // 添加日志记录
      console.log(`[Inventory] 成功获取装备详情: ${response.item.name} (${itemId})`)
    } catch (error) {
      console.error('获取装备详情时发生错误:', error)
      message.error('获取装备详情时发生错误: ' + error.message)
      // 添加错误日志记录
      console.error(`[Inventory] 获取装备详情时发生错误: ${itemId}, 错误: ${error.message}`)
    }
  }

  // 装备详情相关
  const showEquipmentDetailModal = ref(false)
  const selectedEquipment = ref(null)

  // 强化确认弹窗
  const showEnhanceConfirm = ref(false)

  // 强化装备
  const handleEnhanceEquipment = async () => {
    if (!selectedEquipment.value) return
    
    const token = getAuthToken()
    try {
      // 直接从后端获取最新的玩家数据，包括强化石数量
      const userData = await APIService.getUser(token)
      inventoryStore.reinforceStones = userData.reinforceStones || 0
      
      const result = await APIService.enhanceEquipment(token, selectedEquipment.value.id, inventoryStore.reinforceStones)
      
      if (result.success) {
        inventoryStore.reinforceStones -= result.cost
        selectedEquipment.value.stats = { ...result.newStats }
        selectedEquipment.value.enhanceLevel = result.newLevel
        // 更新装备的境界要求
        if (result.newRequiredRealm !== undefined) {
          selectedEquipment.value.requiredRealm = result.newRequiredRealm
        }
        message.success('强化成功')
      } else {
        message.error(result.message || '强化失败')
      }
    } catch (error) {
      console.error('装备强化失败:', error)
      message.error('装备强化失败: ' + error.message)
    }
  }

  // 洗练确认弹窗
  const showReforgeConfirm = ref(false)
  const reforgeResult = ref(null)

  // 洗练装备
  const handleReforgeEquipment = async () => {
    if (!selectedEquipment.value) return
    
    const token = getAuthToken()
    try {
      const result = await APIService.reforgeEquipment(token, selectedEquipment.value.id, playerInfoStore.refinementStones)
      
      if (result.success) {
        playerInfoStore.refinementStones -= result.cost
        reforgeResult.value = result
        showReforgeConfirm.value = true
      } else {
        message.error(result.message || '洗练失败')
      }
    } catch (error) {
      console.error('装备洗练失败:', error)
      message.error('装备洗练失败: ' + error.message)
    }
  }

  // 确认洗练结果
  const confirmReforgeResult = async confirm => {
    if (!reforgeResult.value) return
    
    const token = getAuthToken()
    try {
      const result = await APIService.confirmReforge(
        token, 
        selectedEquipment.value.id, 
        confirm, 
        confirm ? reforgeResult.value.newStats : null
      )
      
      if (result.success) {
        if (confirm) {
          // 用户确认后，应用新属性
          selectedEquipment.value.stats = reforgeResult.value.newStats
          message.success('已确认新属性')
        } else {
          // 用户取消，保留原属性
          message.info('已保留原有属性')
        }
      } else {
        message.error(result.message || '确认洗练结果失败')
      }
    } catch (error) {
      console.error('确认洗练结果失败:', error)
      message.error('确认洗练结果失败: ' + error.message)
    }
    
    showReforgeConfirm.value = false
    reforgeResult.value = null
  }

  // 使用物品
  const useItem = async item => {
    if (item.type === 'pill') {
      const result = inventoryStore.useItem(item, pillsStore, petsStore, playerInfoStore, persistenceStore, combatStore)
      if (result.success) {
        message.success(result.message)
      } else {
        message.error(result.message || '操作失败')
      }
    } else if (item.type === 'pet') {
      // 处理灵宠出战/召回
      const token = getAuthToken();
      
      // 打印操作前的玩家属性
      console.log('[Inventory] 灵宠操作前的玩家属性:', {
        baseAttributes: playerInfoStore.baseAttributes,
        combatAttributes: playerInfoStore.combatAttributes,
        combatResistance: playerInfoStore.combatResistance,
        specialAttributes: playerInfoStore.specialAttributes
      });
      
      
      if (playerInfoStore.activePet?.id === item.id) {
        // 召回灵宠
        console.log(`[Pet Action] 尝试召回灵宠: ${item.name}(${item.id}), 当前activePet:`, playerInfoStore.activePet);
        try {
          const response = await APIService.recallPet(token, item.id);
          console.log(`[Pet Action] 召回灵宠API响应:`, response);
          if (response.success) {
            message.success('召回成功');
            // 更新玩家数据
            // 清除当前出战的灵宠
            playerInfoStore.activePet = response.pet || null;
            console.log(`[Pet Action] 召回成功后activePet状态:`, playerInfoStore.activePet);
            
            // 确保被召回的灵宠对象的isActive属性也被更新
            const recalledPet = inventoryStore.items.find(i => i.type === 'pet' && i.id === item.id);
            if (recalledPet) {
              recalledPet.isActive = false;
              console.log(`[Pet Action] 更新inventoryStore中灵宠isActive为false: ${recalledPet.name}(${recalledPet.id})`);
            }
            
            // 同时更新petsStore中的灵宠数据
            const petInStore = petsStore.pets.find(p => p.id === item.id);
            if (petInStore) {
              petInStore.isActive = false;
              console.log(`[Pet Action] 更新petsStore中灵宠isActive为false: ${petInStore.name}(${petInStore.id})`);
            }
            
            // 重新加载玩家数据以获取更新后的属性
            try {
              const playerDataResponse = await APIService.initializePlayer(token);
              console.log('[Pet Action] 获取更新后的玩家数据:', playerDataResponse);
              
              // 更新玩家属性
              if (playerDataResponse.user) {
                Object.assign(playerInfoStore, playerDataResponse.user);
                console.log('[Pet Action] 已更新玩家属性');
              }
            } catch (error) {
              console.error('[Pet Action] 获取更新后的玩家数据失败:', error);
            }
            
            // 重新加载灵宠列表以确保数据同步
            await loadPetList();
            console.log(`[Pet Action] 召回操作完成，重新加载灵宠列表`);
          } else {
            message.error(response.message || '召回失败');
          }
        } catch (error) {
          console.error('召回灵宠失败:', error);
          message.error('召回灵宠失败: ' + error.message);
        }
      } else {
        // 出战灵宠
        console.log(`[Pet Action] 尝试出战灵宠: ${item.name}(${item.id}), 当前activePet:`, playerInfoStore.activePet);
        try {
          const response = await APIService.deployPet(token, item.id);
          console.log(`[Pet Action] 出战灵宠API响应:`, response);
          if (response.success) {
            message.success('出战成功');
            // 更新玩家数据
            // 更新当前出战的灵宠
            playerInfoStore.activePet = response.pet || item;
            console.log(`[Pet Action] 出战成功后activePet状态:`, playerInfoStore.activePet);
            
            // 确保出战的灵宠对象的isActive属性也被更新
            const deployedPet = inventoryStore.items.find(i => i.type === 'pet' && i.id === item.id);
            if (deployedPet) {
              deployedPet.isActive = true;
              console.log(`[Pet Action] 更新inventoryStore中灵宠isActive为true: ${deployedPet.name}(${deployedPet.id})`);
            }
            
            // 同时更新petsStore中的灵宠数据
            const petInStore = petsStore.pets.find(p => p.id === item.id);
            if (petInStore) {
              petInStore.isActive = true;
              console.log(`[Pet Action] 更新petsStore中灵宠isActive为true: ${petInStore.name}(${petInStore.id})`);
            }
            
            // 重新加载玩家数据以获取更新后的属性
            try {
              const playerDataResponse = await APIService.initializePlayer(token);
              console.log('[Pet Action] 获取更新后的玩家数据:', playerDataResponse);
              
              // 更新玩家属性
              if (playerDataResponse.user) {
                Object.assign(playerInfoStore, playerDataResponse.user);
                console.log('[Pet Action] 已更新玩家属性');
              }
            } catch (error) {
              console.error('[Pet Action] 获取更新后的玩家数据失败:', error);
            }
            
            // 重新加载灵宠列表以确保数据同步
            await loadPetList();
            console.log(`[Pet Action] 出战操作完成，重新加载灵宠列表`);
          } else {
            message.error(response.message || '出战失败');
          }
        } catch (error) {
          console.error('出战灵宠失败:', error);
          message.error('出战灵宠失败: ' + error.message);
        }
      }
      
      // 打印操作后的玩家属性
      console.log('[Inventory] 灵宠操作后的玩家属性:', {
        baseAttributes: playerInfoStore.baseAttributes,
        combatAttributes: playerInfoStore.combatAttributes,
        combatResistance: playerInfoStore.combatResistance,
        specialAttributes: playerInfoStore.specialAttributes
      });
    }
  }

  // 装备属性对比计算
  const equipmentComparison = computed(() => {
    if (!selectedEquipment.value || !selectedEquipmentType.value) return null
    const currentEquipment = equipmentStore.equippedArtifacts[selectedEquipmentType.value]
    if (!currentEquipment) return null
    const comparison = {}
    const allStats = new Set([
      ...Object.keys(selectedEquipment.value.stats || {}), 
      ...Object.keys(currentEquipment.stats || {})
    ])
    allStats.forEach(stat => {
      const selectedValue = (selectedEquipment.value.stats && selectedEquipment.value.stats[stat]) || 0
      const currentValue = (currentEquipment.stats && currentEquipment.stats[stat]) || 0
      const diff = selectedValue - currentValue
      comparison[stat] = {
        current: currentValue,
        selected: selectedValue,
        diff: diff,
        isPositive: diff > 0
      }
    })
    return comparison
  })

  const options = [
    { label: '全部品阶', value: 'all' },
    // 使用统一的灵宠品质配置
    { label: petRarities.mythic.name, value: 'mythic' },
    { label: petRarities.legendary.name, value: 'legendary' },
    { label: petRarities.epic.name, value: 'epic' },
    { label: petRarities.rare.name, value: 'rare' },
    { label: petRarities.uncommon.name, value: 'uncommon' },
    { label: petRarities.common.name, value: 'common' }
  ]
  
  // 装备列表分页
  const currentEquipmentPage = ref(1)
  const equipmentPageSize = ref(8)
  
  // 分页后的装备列表
  const pagedEquipmentList = computed(() => {
    console.log('[Inventory] 计算装备列表分页')
    const start = (currentEquipmentPage.value - 1) * equipmentPageSize.value
    const end = start + equipmentPageSize.value
    const list = filteredEquipmentList.value.slice(start, end)
    console.log(`[Inventory] 装备列表分页结果，起始: ${start}, 结束: ${end}, 数量: ${list.length}`)
    return list
  })
  
  // 装备品质筛选
  const selectedQuality = ref('all')
  
  // 页大小改变处理
  const onEquipmentPageSizeChange = size => {
    equipmentPageSize.value = size
    currentEquipmentPage.value = 1
  }
  
  // 批量释放宠物
  const batchReleasePets = async () => {
    const rarity = selectedRarityToRelease.value === 'all' ? null : selectedRarityToRelease.value
    const token = getAuthToken()
    try {
      const response = await APIService.batchReleasePets(token, { rarity })
      if (response.success) {
        message.success('批量放生成功')
        showBatchReleaseConfirm.value = false
        // 刷新灵宠列表
        await loadPetList()
      } else {
        message.error(response.message || '批量放生失败')
      }
    } catch (error) {
      console.error('批量放生灵宠失败:', error)
      message.error('批量放生灵宠失败: ' + error.message)
    }
  }
  
  // 灵草分组
  const groupedHerbs = computed(() => {
    const herbs = {}
    inventoryStore.items
      .filter(item => item.type === 'herb')
      .forEach(herb => {
        if (herbs[herb.id]) {
          herbs[herb.id].count++
        } else {
          herbs[herb.id] = { ...herb, count: 1 }
        }
      })
    return Object.values(herbs)
  })
  
  // 丹药分组
  const groupedPills = computed(() => {
    const pills = {}
    inventoryStore.items
      .filter(item => item.type === 'pill')
      .forEach(pill => {
        if (pills[pill.id]) {
          pills[pill.id].count++
        } else {
          pills[pill.id] = { ...pill, count: 1 }
        }
      })
    return Object.values(pills)
  })
  
  // 丹方分组
  const groupedFormulas = computed(() => {
    const complete = []
    const incomplete = []
    
    inventoryStore.items
      .filter(item => item.type === 'formula')
      .forEach(formula => {
        if (formula.fragments >= formula.fragmentsNeeded) {
          complete.push(formula)
        } else {
          incomplete.push(formula)
        }
      })
      
    return { complete, incomplete }
  })
  
 
</script>

<style scoped>
  .n-card {
    cursor: pointer;
  }

  .reforge-compare {
    display: flex;
    justify-content: space-between;
    gap: 20px;
    margin: 16px 0;
  }

  .old-stats,
  .new-stats {
    flex: 1;
    padding: 16px;
    border-radius: 8px;
    background-color: rgba(0, 0, 0, 0.05);
  }

  .old-stats h3,
  .new-stats h3 {
    margin-top: 0;
    margin-bottom: 12px;
    font-size: 16px;
    color: #666;
  }
</style>
