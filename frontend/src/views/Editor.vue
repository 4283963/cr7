<template>
  <div class="editor-page">
    <div class="editor-header">
      <div class="header-left">
        <button class="back-btn" @click="goBack">
          <span>←</span>
          <span>返回</span>
        </button>
        <div class="script-info">
          <input 
            v-if="isEditingName"
            ref="nameInput"
            type="text" 
            v-model="editName"
            class="script-name-input"
            @blur="saveName"
            @keyup.enter="saveName"
            @keyup.esc="cancelEditName"
          />
          <h1 v-else class="script-name" @dblclick="startEditName">
            {{ script?.name || '未命名剧本' }}
          </h1>
          <span class="script-desc">{{ script?.description || '' }}</span>
        </div>
      </div>
      <div class="header-right">
        <span class="status-indicator" :class="{ saved: isSaved }">
          <span class="status-dot"></span>
          {{ isSaved ? '已保存' : '未保存' }}
        </span>
        <button class="btn-secondary" @click="saveScript">
          💾 保存
        </button>
      </div>
    </div>

    <div class="editor-body">
      <div class="editor-left">
        <div class="stage-wrapper">
          <ShadowPuppet 
            :view-width="600" 
            :view-height="450"
            @bone-select="onBoneSelect"
            @angle-change="onAngleChange"
          />
        </div>
        
        <div class="bone-list-panel">
          <div class="panel-header">
            <span class="panel-title">骨骼列表</span>
            <span class="panel-count">{{ bones.length }} 个关节</span>
          </div>
          <div class="bone-list">
            <div 
              v-for="bone in bones" 
              :key="bone.id"
              class="bone-item"
              :class="{ selected: selectedBoneId === bone.id }"
              @click="selectBone(bone.id)"
            >
              <span class="bone-color-dot" :style="{ background: bone.color }"></span>
              <span class="bone-name">{{ bone.name }}</span>
              <span class="bone-angle">{{ getBoneAngleDisplay(bone.id) }}°</span>
            </div>
          </div>
        </div>
      </div>

      <div class="editor-right">
        <div class="properties-panel">
          <div class="panel-header">
            <span class="panel-title">剧本属性</span>
          </div>
          <div class="properties-content">
            <div class="prop-row">
              <label>总时长 (秒)</label>
              <input 
                type="number" 
                :value="duration"
                @change="onDurationChange"
                min="1"
                max="300"
                class="prop-input"
              />
            </div>
            <div class="prop-row">
              <label>帧率 (FPS)</label>
              <span class="prop-value">{{ fps }}</span>
            </div>
            <div class="prop-row">
              <label>轨道数</label>
              <span class="prop-value">{{ tracks.length }}</span>
            </div>
          </div>
        </div>

        <div class="shortcuts-panel">
          <div class="panel-header">
            <span class="panel-title">快捷操作</span>
          </div>
          <div class="shortcuts-content">
            <div class="shortcut-item">
              <span class="shortcut-key">空格</span>
              <span class="shortcut-desc">播放/暂停</span>
            </div>
            <div class="shortcut-item">
              <span class="shortcut-key">←/→</span>
              <span class="shortcut-desc">逐帧调整</span>
            </div>
            <div class="shortcut-item">
              <span class="shortcut-key">Ctrl+滚轮</span>
              <span class="shortcut-desc">时间轴缩放</span>
            </div>
            <div class="shortcut-item">
              <span class="shortcut-key">双击</span>
              <span class="shortcut-desc">添加关键帧</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="editor-timeline">
      <Timeline 
        @time-change="onTimeChange"
        @keyframe-select="onKeyframeSelect"
        @track-select="onTrackSelect"
      />
    </div>

    <div v-if="loading" class="loading-overlay">
      <div class="loading-spinner"></div>
      <p>加载中...</p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import ShadowPuppet from '@/components/ShadowPuppet.vue'
import Timeline from '@/components/Timeline.vue'
import { useEditorStore } from '@/stores/editor'
import { formatTime } from '@/utils/animation'

const route = useRoute()
const router = useRouter()
const store = useEditorStore()

const { selectedBoneId, bones, tracks, duration, fps, isPlaying, script } = storeToRefs(store)

const loading = ref(true)
const isSaved = ref(true)
const isEditingName = ref(false)
const editName = ref('')
const nameInput = ref(null)

onMounted(async () => {
  const scriptId = route.params.id
  if (scriptId) {
    try {
      await store.loadScript(scriptId)
      isSaved.value = true
    } catch (e) {
      console.error('Failed to load script:', e)
    }
  }
  loading.value = false
  
  window.addEventListener('keydown', onKeyDown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeyDown)
  store.pause()
})

watch(() => route.params.id, async (newId) => {
  if (newId) {
    loading.value = true
    try {
      await store.loadScript(newId)
      isSaved.value = true
    } catch (e) {
      console.error('Failed to load script:', e)
    }
    loading.value = false
  }
})

function onKeyDown(e) {
  if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return
  
  if (e.code === 'Space') {
    e.preventDefault()
    store.togglePlay()
  } else if (e.code === 'ArrowLeft') {
    e.preventDefault()
    store.setTime(Math.max(0, store.currentTime - 1 / fps.value))
  } else if (e.code === 'ArrowRight') {
    e.preventDefault()
    store.setTime(Math.min(duration.value, store.currentTime + 1 / fps.value))
  }
}

function goBack() {
  router.push('/')
}

function startEditName() {
  editName.value = script.value?.name || ''
  isEditingName.value = true
  nextTick(() => {
    nameInput.value?.focus()
    nameInput.value?.select()
  })
}

function saveName() {
  if (script.value && editName.value.trim()) {
    script.value.name = editName.value.trim()
    isSaved.value = false
  }
  isEditingName.value = false
}

function cancelEditName() {
  isEditingName.value = false
}

async function saveScript() {
  try {
    await store.saveScript()
    isSaved.value = true
  } catch (e) {
    console.error('Save failed:', e)
  }
}

function selectBone(boneId) {
  store.selectBone(boneId)
}

function onBoneSelect(boneId) {
}

function onAngleChange({ boneId, angle }) {
  isSaved.value = false
}

function onTimeChange(time) {
}

function onKeyframeSelect(info) {
}

function onTrackSelect(index) {
}

function onDurationChange(event) {
  const val = parseFloat(event.target.value)
  if (!isNaN(val) && val > 0) {
    store.setDuration(val)
    isSaved.value = false
  }
}

function getBoneAngleDisplay(boneId) {
  const angle = store.getBoneAngle(boneId)
  return angle.toFixed(1)
}
</script>

<style scoped>
.editor-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  background: #252545;
  border-bottom: 1px solid #3a3a5a;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  background: #3a3a5a;
  color: #c0c0e0;
  border-radius: 4px;
  font-size: 13px;
  transition: all 0.2s;
}

.back-btn:hover {
  background: #4a4a7a;
  color: #e0e0e0;
}

.script-info {
  display: flex;
  flex-direction: column;
}

.script-name {
  font-size: 18px;
  font-weight: 600;
  color: #e8b059;
  cursor: text;
}

.script-name-input {
  font-size: 18px;
  font-weight: 600;
  color: #e8b059;
  background: #1a1a2e;
  border: 1px solid #e8b059;
  border-radius: 4px;
  padding: 2px 8px;
  outline: none;
}

.script-desc {
  font-size: 12px;
  color: #8080a0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #e8b059;
}

.status-indicator.saved {
  color: #6bcb77;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #e8b059;
}

.status-indicator.saved .status-dot {
  background: #6bcb77;
}

.btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: #3a3a5a;
  color: #e0e0e0;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s;
}

.btn-secondary:hover {
  background: #4a4a7a;
}

.editor-body {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.editor-left {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 16px;
  gap: 16px;
  overflow: hidden;
}

.stage-wrapper {
  flex: 1;
  min-height: 0;
}

.bone-list-panel {
  height: 200px;
  flex-shrink: 0;
  background: #252545;
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 14px;
  background: #2d2d50;
  border-bottom: 1px solid #3a3a5a;
}

.panel-title {
  font-size: 13px;
  font-weight: 600;
  color: #c0c0e0;
}

.panel-count {
  font-size: 11px;
  color: #8080a0;
}

.bone-list {
  flex: 1;
  overflow-y: auto;
  padding: 4px;
}

.bone-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.15s;
}

.bone-item:hover {
  background: #3a3a5a;
}

.bone-item.selected {
  background: rgba(232, 176, 89, 0.15);
}

.bone-color-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.bone-name {
  flex: 1;
  font-size: 13px;
  color: #b0b0d0;
}

.bone-item.selected .bone-name {
  color: #e8b059;
}

.bone-angle {
  font-size: 12px;
  color: #707090;
  font-family: 'Courier New', monospace;
}

.editor-right {
  width: 240px;
  flex-shrink: 0;
  padding: 16px 16px 16px 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.properties-panel,
.shortcuts-panel {
  background: #252545;
  border-radius: 8px;
  overflow: hidden;
}

.properties-content {
  padding: 12px 14px;
}

.prop-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #3a3a5a;
}

.prop-row:last-child {
  border-bottom: none;
}

.prop-row label {
  font-size: 12px;
  color: #9090b0;
}

.prop-value {
  font-size: 13px;
  color: #e0e0e0;
  font-weight: 500;
}

.prop-input {
  width: 80px;
  padding: 4px 8px;
  background: #1a1a2e;
  border: 1px solid #3a3a5a;
  border-radius: 4px;
  color: #e0e0e0;
  font-size: 13px;
  text-align: right;
  outline: none;
}

.prop-input:focus {
  border-color: #e8b059;
}

.shortcuts-content {
  padding: 12px 14px;
}

.shortcut-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 0;
}

.shortcut-key {
  display: inline-block;
  min-width: 60px;
  padding: 3px 8px;
  background: #1a1a2e;
  border: 1px solid #3a3a5a;
  border-radius: 4px;
  font-size: 11px;
  color: #a0a0c0;
  text-align: center;
  font-family: 'Courier New', monospace;
}

.shortcut-desc {
  font-size: 12px;
  color: #8080a0;
}

.editor-timeline {
  height: 300px;
  flex-shrink: 0;
  padding: 0 16px 16px;
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(26, 26, 46, 0.9);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #3a3a5a;
  border-top-color: #e8b059;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 12px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-overlay p {
  color: #a0a0c0;
  font-size: 14px;
}
</style>
