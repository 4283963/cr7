<template>
  <div class="audio-panel" v-if="script">
    <div class="panel-header">
      <div class="title">
        <span class="title-icon">♪</span>
        <span class="title-text">鼓点与节拍</span>
      </div>
      <button class="close-btn" @click="$emit('close')" title="关闭">×</button>
    </div>

    <div class="panel-body">

      <div class="section">
        <div class="section-title">音频文件</div>
        <div class="upload-area" v-if="!audioTrack || !audioTrack.fileName">
          <input
            ref="fileInput"
            type="file"
            accept="audio/*"
            class="hidden-input"
            @change="onFileSelect"
          />
          <button class="upload-btn" @click="$refs.fileInput.click()" :disabled="uploading">
            <span class="upload-icon">↑</span>
            <span>{{ uploading ? `上传中 ${Math.round(uploadProgress * 100)}%` : '选择鼓点音频文件' }}</span>
          </button>
          <div class="upload-hint">支持 MP3 / WAV / OGG / M4A · 建议 30 秒以内</div>
        </div>
        <div class="audio-info" v-else>
          <div class="audio-file-info">
            <div class="file-icon">♪</div>
            <div class="file-meta">
              <div class="file-name">{{ audioTrack.fileName }}</div>
              <div class="file-details">
                {{ formatFileSize(audioTrack.fileSize) }} · 
                {{ formatTime(audioTrack.duration || 0) }} · 
                音量: {{ Math.round(audioTrack.volume * 100) }}%
              </div>
            </div>
          </div>
          <div class="volume-row">
            <span class="volume-label">音量</span>
            <input
              type="range"
              min="0"
              max="1"
              step="0.05"
              :value="audioTrack.volume"
              @input="onVolumeChange($event.target.value)"
              class="volume-slider"
            />
            <span class="volume-val">{{ Math.round(audioTrack.volume * 100) }}%</span>
          </div>
          <div class="audio-actions">
            <button class="action-btn" @click="playPreview">
              <span>{{ isPreviewPlaying ? '⏸ 暂停试听' : '▶ 试听' }}</span>
            </button>
            <button class="action-btn" @click="$refs.fileInput.click()">
              <span>⟳ 替换音频</span>
            </button>
            <button class="action-btn danger" @click="onRemoveAudio">
              <span>✕ 删除</span>
            </button>
          </div>
        </div>
      </div>

      <div class="section">
        <div class="section-title">
          <span>节拍点设置</span>
          <span class="beat-count-badge">{{ enabledBeatCount }} / {{ beats.length }}</span>
        </div>

        <div class="bpm-generator">
          <div class="bpm-row">
            <label class="bpm-label">BPM</label>
            <input
              type="number"
              v-model.number="bpmInput"
              min="30"
              max="300"
              step="1"
              class="bpm-input"
            />
            <label class="bpm-label">偏移</label>
            <input
              type="number"
              v-model.number="bpmOffset"
              min="0"
              :max="script.duration"
              step="0.05"
              class="bpm-input small"
            />
            <span class="bpm-unit">秒</span>
            <button class="generate-btn" @click="onGenerateBeats">
              <span>按 BPM 生成节拍</span>
            </button>
          </div>
          <div class="bpm-hint">
            节拍间隔 = 60 ÷ BPM 秒。例：BPM=120 时每 0.5 秒一个鼓点
          </div>
        </div>

        <div class="tolerance-row">
          <label class="tol-label">卡点容差</label>
          <input
            type="range"
            min="0.01"
            max="0.3"
            step="0.005"
            :value="audioTrack ? audioTrack.syncTolerance : 0.08"
            @input="onToleranceChange($event.target.value)"
            class="tol-slider"
          />
          <span class="tol-value">{{ Math.round((audioTrack ? audioTrack.syncTolerance : 0.08) * 1000) }} ms</span>
        </div>

        <div class="start-offset-row">
          <label class="tol-label">起始偏移</label>
          <input
            type="number"
            :value="audioTrack ? audioTrack.startTime : 0"
            @input="onStartTimeChange($event.target.value)"
            min="0"
            step="0.05"
            class="start-input"
          />
          <span class="tol-unit">秒</span>
          <div class="hint-inline">音乐相对动作提前启动的时间</div>
        </div>
      </div>

      <div class="section">
        <div class="section-title">
          <span>节拍列表</span>
          <div class="list-actions">
            <button class="mini-btn" @click="addCurrentAsBeat" :disabled="!audioTrack">+ 在当前时间</button>
            <button class="mini-btn danger" @click="clearAllBeats" :disabled="beats.length === 0">清空</button>
          </div>
        </div>

        <div class="beats-list" ref="beatsList">
          <div
            v-for="(beat, idx) in beats"
            :key="idx"
            class="beat-row"
            :class="{ enabled: beat.enabled, synced: hasSyncedAtBeat(beat.time) }"
          >
            <input
              type="checkbox"
              :checked="beat.enabled"
              @change="toggleBeatEnabled(idx)"
              class="beat-checkbox"
            />
            <span class="beat-index">{{ idx + 1 }}</span>
            <input
              type="number"
              :value="beat.time"
              @change="onBeatTimeChange(idx, $event.target.value)"
              step="0.05"
              min="0"
              class="beat-time-input"
            />
            <span class="beat-time-unit">秒</span>
            <input
              type="text"
              v-model="beat.label"
              placeholder="标签(选填)"
              maxlength="6"
              class="beat-label-input"
              @change="saveBeatsNow"
            />
            <span class="beat-sync-tag" v-if="hasSyncedAtBeat(beat.time)">
              ✓ {{ getSyncedCountAtBeat(beat.time) }} 卡
            </span>
            <button class="beat-delete-btn" @click="onDeleteBeat(idx)">×</button>
          </div>
          <div v-if="beats.length === 0" class="empty-list">
            暂无节拍。上传音频后用 BPM 自动生成，或在时间轴音频轨道双击手动添加。
          </div>
        </div>
      </div>
    </div>

    <div class="panel-footer">
      <div class="sync-summary" v-if="audioTrack && beats.length > 0">
        <span>卡点统计：</span>
        <span class="sync-num">{{ syncedBeatsCount }} / {{ enabledBeatCount }}</span>
        <span> 节拍已被动作命中</span>
        <span class="sync-rate">({{ Math.round(syncedBeatsCount / Math.max(1, enabledBeatCount) * 100) }}%)</span>
      </div>
      <div class="footer-actions">
        <button class="btn-secondary" @click="$emit('close')">关闭</button>
        <button class="btn-primary" @click="onSaveAndClose" :disabled="saving">
          {{ saving ? '保存中...' : '保存节拍设置' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onUnmounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useEditorStore } from '@/stores/editor'

const emit = defineEmits(['close'])

const store = useEditorStore()
const { script, audioTrack, beats, syncTolerance, currentTime } = storeToRefs(store)

const fileInput = ref(null)
const uploading = ref(false)
const uploadProgress = ref(0)
const saving = ref(false)
const bpmInput = ref(100)
const bpmOffset = ref(0)
const isPreviewPlaying = ref(false)
const previewAudio = ref(null)
const beatsList = ref(null)

const enabledBeatCount = computed(() => beats.value.filter(b => b.enabled).length)

const syncedBeatsCount = computed(() => {
  const tol = syncTolerance.value
  const hitBeats = new Set()
  if (!script.value?.tracks) return 0
  script.value.tracks.forEach(track => {
    if (!track.keyframes) return
    track.keyframes.forEach(kf => {
      beats.value.forEach((beat, bi) => {
        if (!beat.enabled) return
        if (Math.abs(kf.time - beat.time) <= tol) {
          hitBeats.add(bi)
        }
      })
    })
  })
  return hitBeats.size
})

function formatTime(s) {
  if (!s || s < 0) s = 0
  const m = Math.floor(s / 60)
  const sec = (s % 60)
  return `${m}:${sec.toFixed(sec % 1 === 0 ? 0 : 2).padStart(m > 0 ? 5 : 4, '0')}`
}

function formatFileSize(bytes) {
  if (!bytes) return '0 B'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1024 / 1024).toFixed(2) + ' MB'
}

async function onFileSelect(e) {
  const file = e.target.files?.[0]
  if (!file) return
  uploading.value = true
  uploadProgress.value = 0
  try {
    await store.uploadAudioFile(file, (p) => { uploadProgress.value = p })
  } catch (err) {
    alert('上传失败：' + err.message)
  } finally {
    uploading.value = false
    if (fileInput.value) fileInput.value.value = ''
  }
}

async function onRemoveAudio() {
  if (!confirm('确定要移除音频和所有节拍吗？')) return
  try {
    await store.removeAudio()
  } catch (e) {
    alert('删除失败')
  }
}

function onVolumeChange(v) {
  if (!script.value?.audioTrack) return
  script.value.audioTrack.volume = Number(v)
  saveBeatsNow()
}

function onToleranceChange(v) {
  if (!script.value?.audioTrack) return
  script.value.audioTrack.syncTolerance = Number(v)
  store.checkAllKeyframesSynced()
}

function onStartTimeChange(v) {
  if (!script.value?.audioTrack) return
  script.value.audioTrack.startTime = Number(v)
}

async function onGenerateBeats() {
  if (!audioTrack.value) return
  saving.value = true
  try {
    await store.generateBeatsFromBPM(bpmInput.value, bpmOffset.value)
  } catch (e) {
    alert('生成失败')
  } finally {
    saving.value = false
  }
}

function addCurrentAsBeat() {
  if (!audioTrack.value) return
  store.addBeat(Math.round(currentTime.value * 100) / 100)
  saveBeatsNow()
}

function clearAllBeats() {
  if (!audioTrack.value || beats.value.length === 0) return
  if (!confirm(`确定要清空 ${beats.value.length} 个节拍吗？`)) return
  script.value.audioTrack.beats = []
  store.checkAllKeyframesSynced()
}

function toggleBeatEnabled(idx) {
  if (!script.value?.audioTrack?.beats[idx]) return
  script.value.audioTrack.beats[idx].enabled = !script.value.audioTrack.beats[idx].enabled
  store.checkAllKeyframesSynced()
}

function onBeatTimeChange(idx, val) {
  if (!script.value?.audioTrack?.beats[idx]) return
  script.value.audioTrack.beats[idx].time = Number(val)
  script.value.audioTrack.beats.sort((a, b) => a.time - b.time)
  store.checkAllKeyframesSynced()
  saveBeatsNow()
}

function onDeleteBeat(idx) {
  store.removeBeat(idx)
  saveBeatsNow()
}

function hasSyncedAtBeat(beatTime) {
  const tol = syncTolerance.value
  if (!script.value?.tracks) return false
  for (const track of script.value.tracks) {
    if (!track.keyframes) continue
    for (const kf of track.keyframes) {
      if (Math.abs(kf.time - beatTime) <= tol) return true
    }
  }
  return false
}

function getSyncedCountAtBeat(beatTime) {
  let n = 0
  const tol = syncTolerance.value
  if (!script.value?.tracks) return 0
  for (const track of script.value.tracks) {
    if (!track.keyframes) continue
    for (const kf of track.keyframes) {
      if (Math.abs(kf.time - beatTime) <= tol) n++
    }
  }
  return n
}

async function saveBeatsNow() {
  if (!script.value?.audioTrack?.beats) return
  try {
    await store.saveBeats(script.value.audioTrack.beats, {
      syncTolerance: script.value.audioTrack.syncTolerance,
      startTime: script.value.audioTrack.startTime,
      volume: script.value.audioTrack.volume
    })
  } catch (e) {
    console.warn('save beats failed', e)
  }
}

async function onSaveAndClose() {
  saving.value = true
  try {
    if (script.value?.audioTrack?.beats) {
      await store.saveBeats(script.value.audioTrack.beats, {
        syncTolerance: script.value.audioTrack.syncTolerance,
        startTime: script.value.audioTrack.startTime,
        volume: script.value.audioTrack.volume
      })
    }
    emit('close')
  } catch (e) {
    alert('保存失败：' + e.message)
  } finally {
    saving.value = false
  }
}

function playPreview() {
  if (!audioTrack.value?.fileName || !store.audioUrl) return
  if (previewAudio.value) {
    if (isPreviewPlaying.value) {
      previewAudio.value.pause()
      isPreviewPlaying.value = false
    } else {
      previewAudio.value.play().catch(() => {})
      isPreviewPlaying.value = true
    }
    return
  }
  previewAudio.value = new Audio(store.audioUrl)
  previewAudio.value.volume = audioTrack.value.volume
  previewAudio.value.onended = () => { isPreviewPlaying.value = false }
  previewAudio.value.play().then(() => {
    isPreviewPlaying.value = true
  }).catch(() => {})
}

watch(
  () => audioTrack.value?.beats?.length,
  () => {
    setTimeout(() => {
      if (beatsList.value) beatsList.value.scrollTop = 0
    }, 50)
  }
)

onUnmounted(() => {
  if (previewAudio.value) {
    previewAudio.value.pause()
    previewAudio.value = null
  }
})
</script>

<style scoped>
.audio-panel {
  width: 440px;
  height: 100%;
  background: #1e1e38;
  border-left: 1px solid #3a3a5a;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  border-bottom: 1px solid #3a3a5a;
  background: #252545;
  flex-shrink: 0;
}

.title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #e0e0e0;
  font-size: 14px;
}

.title-icon {
  font-size: 18px;
  color: #4ade80;
}

.close-btn {
  width: 28px;
  height: 28px;
  background: transparent;
  color: #8080a0;
  border-radius: 4px;
  font-size: 18px;
  line-height: 1;
  transition: all 0.15s;
}

.close-btn:hover {
  background: #3a3a5a;
  color: #e0e0e0;
}

.panel-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px;
}

.section {
  margin-bottom: 20px;
}

.section-title {
  font-size: 12px;
  font-weight: 600;
  color: #8080a0;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 10px;
  padding-bottom: 6px;
  border-bottom: 1px solid #3a3a5a;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.beat-count-badge {
  background: rgba(74, 222, 128, 0.15);
  color: #4ade80;
  font-size: 10px;
  padding: 2px 8px;
  border-radius: 10px;
}

.list-actions {
  display: flex;
  gap: 6px;
}

.mini-btn {
  font-size: 10px;
  padding: 3px 8px;
  background: #3a3a5a;
  color: #c0c0e0;
  border-radius: 3px;
  transition: all 0.15s;
}

.mini-btn:hover:not(:disabled) {
  background: #4a4a7a;
}

.mini-btn.danger {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.mini-btn.danger:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.3);
}

.mini-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.hidden-input {
  display: none;
}

.upload-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 24px;
  border: 2px dashed #3a3a5a;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.02);
}

.upload-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  background: linear-gradient(135deg, rgba(74, 222, 128, 0.2), rgba(74, 222, 128, 0.08));
  border: 1px solid rgba(74, 222, 128, 0.3);
  color: #4ade80;
  border-radius: 6px;
  font-weight: 500;
  transition: all 0.2s;
}

.upload-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, rgba(74, 222, 128, 0.3), rgba(74, 222, 128, 0.15));
  transform: translateY(-1px);
}

.upload-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.upload-icon {
  font-size: 18px;
}

.upload-hint {
  font-size: 10px;
  color: #707090;
}

.audio-info {
  background: rgba(74, 222, 128, 0.05);
  border: 1px solid rgba(74, 222, 128, 0.2);
  border-radius: 8px;
  padding: 12px;
}

.audio-file-info {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.file-icon {
  width: 44px;
  height: 44px;
  background: rgba(74, 222, 128, 0.15);
  color: #4ade80;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
}

.file-meta {
  flex: 1;
  min-width: 0;
}

.file-name {
  font-size: 13px;
  font-weight: 500;
  color: #c0c0e0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-details {
  font-size: 10px;
  color: #8080a0;
  margin-top: 2px;
}

.volume-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.volume-label {
  font-size: 11px;
  color: #8080a0;
  min-width: 32px;
}

.volume-slider {
  flex: 1;
  accent-color: #4ade80;
}

.volume-val {
  font-size: 10px;
  color: #8080a0;
  min-width: 36px;
  text-align: right;
}

.audio-actions {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.action-btn {
  flex: 1;
  padding: 6px 10px;
  background: #3a3a5a;
  color: #c0c0e0;
  border-radius: 4px;
  font-size: 11px;
  transition: all 0.15s;
  min-width: 70px;
}

.action-btn:hover {
  background: #4a4a7a;
  color: #e0e0e0;
}

.action-btn.danger {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.action-btn.danger:hover {
  background: rgba(239, 68, 68, 0.3);
}

.bpm-generator {
  background: rgba(0, 0, 0, 0.15);
  border-radius: 6px;
  padding: 10px;
  margin-bottom: 12px;
}

.bpm-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.bpm-label {
  font-size: 11px;
  color: #8080a0;
}

.bpm-input {
  width: 64px;
  padding: 4px 8px;
  background: #2a2a4a;
  border: 1px solid #3a3a5a;
  border-radius: 4px;
  color: #e0e0e0;
  font-size: 12px;
}

.bpm-input.small {
  width: 56px;
}

.bpm-unit {
  font-size: 10px;
  color: #707090;
}

.generate-btn {
  margin-left: auto;
  padding: 6px 12px;
  background: rgba(74, 222, 128, 0.2);
  color: #4ade80;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  transition: all 0.15s;
  border: 1px solid rgba(74, 222, 128, 0.3);
}

.generate-btn:hover {
  background: rgba(74, 222, 128, 0.35);
}

.bpm-hint {
  font-size: 10px;
  color: #707090;
  margin-top: 8px;
}

.tolerance-row,
.start-offset-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.tol-label {
  font-size: 11px;
  color: #8080a0;
  min-width: 56px;
}

.tol-slider {
  flex: 1;
  accent-color: #4ade80;
}

.tol-value,
.tol-unit {
  font-size: 10px;
  color: #8080a0;
  min-width: 44px;
  text-align: right;
}

.hint-inline {
  font-size: 10px;
  color: #707090;
  margin-left: auto;
}

.start-input {
  width: 64px;
  padding: 4px 8px;
  background: #2a2a4a;
  border: 1px solid #3a3a5a;
  border-radius: 4px;
  color: #e0e0e0;
  font-size: 12px;
}

.beats-list {
  max-height: 220px;
  overflow-y: auto;
  background: rgba(0, 0, 0, 0.15);
  border-radius: 6px;
  border: 1px solid #2a2a4a;
}

.beat-row {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 8px;
  border-bottom: 1px solid rgba(58, 58, 90, 0.5);
  transition: background 0.15s;
}

.beat-row:last-child {
  border-bottom: none;
}

.beat-row:hover {
  background: rgba(74, 222, 128, 0.05);
}

.beat-row.enabled {
  background: rgba(74, 222, 128, 0.03);
}

.beat-row.synced {
  border-left: 3px solid #4ade80;
}

.beat-checkbox {
  width: 14px;
  height: 14px;
  accent-color: #4ade80;
  cursor: pointer;
  flex-shrink: 0;
}

.beat-index {
  font-size: 10px;
  color: #707090;
  min-width: 22px;
  text-align: right;
}

.beat-time-input {
  width: 68px;
  padding: 3px 6px;
  background: #1a1a2e;
  border: 1px solid #3a3a5a;
  border-radius: 3px;
  color: #e0e0e0;
  font-size: 11px;
  font-family: monospace;
}

.beat-time-unit {
  font-size: 9px;
  color: #707090;
}

.beat-label-input {
  flex: 1;
  min-width: 0;
  padding: 3px 6px;
  background: #1a1a2e;
  border: 1px solid #3a3a5a;
  border-radius: 3px;
  color: #c0c0e0;
  font-size: 10px;
}

.beat-sync-tag {
  font-size: 10px;
  padding: 1px 6px;
  background: rgba(74, 222, 128, 0.15);
  color: #4ade80;
  border-radius: 3px;
  font-weight: 500;
}

.beat-delete-btn {
  width: 20px;
  height: 20px;
  background: transparent;
  color: #707090;
  border-radius: 3px;
  font-size: 14px;
  line-height: 1;
  flex-shrink: 0;
  transition: all 0.15s;
}

.beat-delete-btn:hover {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

.empty-list {
  padding: 24px 16px;
  text-align: center;
  font-size: 11px;
  color: #707090;
  line-height: 1.6;
}

.panel-footer {
  padding: 12px 16px;
  border-top: 1px solid #3a3a5a;
  background: #252545;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  gap: 12px;
}

.sync-summary {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: #8080a0;
}

.sync-num {
  font-size: 14px;
  font-weight: 700;
  color: #4ade80;
  margin: 0 2px;
}

.sync-rate {
  color: #a0a0c0;
}

.footer-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.btn-secondary {
  padding: 8px 16px;
  background: #3a3a5a;
  color: #c0c0e0;
  border-radius: 4px;
  font-size: 12px;
  transition: all 0.15s;
}

.btn-secondary:hover {
  background: #4a4a7a;
  color: #e0e0e0;
}

.btn-primary {
  padding: 8px 16px;
  background: linear-gradient(135deg, #e8b059, #d4944a);
  color: #1a1a2e;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  transition: all 0.15s;
}

.btn-primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #f4c069, #e8a45a);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
