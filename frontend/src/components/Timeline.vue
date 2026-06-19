<template>
  <div class="timeline-container">
    <div class="timeline-header">
      <div class="timeline-controls">
        <div class="play-controls">
          <button class="play-btn" @click="stop" title="停止">
            <span class="btn-icon">⏹</span>
          </button>
          <button class="play-btn play-main" @click="togglePlay" :title="isPlaying ? '暂停' : '播放'">
            <span class="btn-icon">{{ isPlaying ? '⏸' : '▶' }}</span>
          </button>
        </div>
        
        <div class="time-display">
          <span class="time-current">{{ formatTime(currentTime) }}</span>
          <span class="time-sep">/</span>
          <span class="time-total">{{ formatTime(duration) }}</span>
        </div>

        <div class="audio-status" v-if="audioTrack">
          <div class="audio-status-icon" :class="{ active: isAudioPlaying }">♪</div>
          <div class="audio-status-text">
            <div class="audio-filename">{{ audioTrack.fileName }}</div>
            <div class="audio-beats">节拍: {{ enabledBeatCount }} 个 · 容差: {{ (syncTolerance * 1000).toFixed(0) }}ms</div>
          </div>
          <div class="audio-sync">
            <span class="sync-count" :style="{ color: syncedCount > 0 ? '#4ade80' : '#707090' }">✓ {{ syncedCount }} 卡</span>
          </div>
        </div>
        <div class="audio-status" v-else>
          <div class="audio-status-icon muted">♪</div>
          <div class="audio-status-text muted-text">未添加音频，可在右侧面板上传鼓点</div>
        </div>
      </div>

      <div class="timeline-ruler-container">
        <div class="ruler-track-label">
          <span>时间轴</span>
        </div>
        <div 
          ref="rulerEl"
          class="timeline-ruler"
          @mousedown="onRulerMouseDown"
          @mousemove="onRulerMouseMove"
          @mouseup="onRulerMouseUp"
          @mouseleave="onRulerMouseUp"
        >
          <svg class="ruler-svg" :width="Math.max(rulerWidth, tracksWidth)" :height="rulerHeight">
            <g v-for="tick in ticks" :key="'t'+tick.time">
              <line
                :x1="tick.x"
                y1="0"
                :x2="tick.x"
                :y2="tick.major ? 20 : 10"
                :stroke="tick.major ? '#6a6a9a' : '#4a4a7a'"
                stroke-width="1"
              />
              <text
                v-if="tick.major"
                :x="tick.x + 3"
                y="28"
                fill="#a0a0c0"
                font-size="10"
              >
                {{ formatTickTime(tick.time) }}
              </text>
            </g>

            <g v-for="(beat, idx) in beats" :key="'br'+idx" v-if="beat.enabled">
              <line
                :x1="beat.time * pixelsPerSecond"
                y1="0"
                :x2="beat.time * pixelsPerSecond"
                :y2="rulerHeight"
                :stroke="beatFlash === beat.time ? '#4ade80' : 'rgba(74, 222, 128, 0.3)'"
                stroke-width="1"
                stroke-dasharray="3,2"
              />
            </g>
          </svg>
          
          <div 
            class="playhead"
            :style="{ left: playheadX + 'px' }"
            @mousedown.stop="onPlayheadMouseDown"
          >
            <div class="playhead-triangle"></div>
            <div class="playhead-line"></div>
          </div>
        </div>
      </div>
    </div>

    <div class="timeline-body" ref="bodyEl">
      <div class="track-labels">
        <div 
          v-for="(track, index) in tracks" 
          :key="track.boneId"
          class="track-label"
          :class="{ selected: selectedTrackIndex === index }"
          @click="selectTrack(index)"
        >
          <span class="track-name">{{ track.boneName }}</span>
          <span class="track-bone-id">{{ track.boneId }}</span>
        </div>
        <div class="track-label audio-label" :class="{ 'has-audio': !!audioTrack }" @click="$emit('openAudioPanel')">
          <span class="track-name">
            <span class="audio-icon">♪</span>
            {{ audioTrack ? '鼓点伴奏' : '音频轨道' }}
          </span>
          <span class="track-bone-id">{{ audioTrack ? '点击管理' : '点击添加' }}</span>
        </div>
      </div>

      <div 
        ref="tracksEl"
        class="tracks-container"
        @mousedown="onTracksMouseDown"
        @mousemove="onTracksMouseMove"
        @mouseup="onTracksMouseUp"
        @mouseleave="onTracksMouseUp"
        @wheel="onTracksWheel"
      >
        <div 
          class="tracks-content"
          :style="{ width: tracksWidth + 'px', height: totalContentHeight + 'px' }"
        >
          <div 
            v-for="(track, trackIdx) in tracks" 
            :key="track.boneId"
            class="track-row"
            :class="{ selected: selectedTrackIndex === trackIdx }"
            :style="{ top: trackIdx * trackHeight + 'px' }"
          >
            <svg 
              class="track-svg"
              :width="tracksWidth"
              :height="trackHeight"
            >
              <line
                x1="0"
                :y1="trackHeight - 1"
                :x2="tracksWidth"
                :y2="trackHeight - 1"
                stroke="#3a3a5a"
                stroke-width="1"
              />

              <g v-for="(beat, bidx) in beats" :key="'bt'+bidx" v-if="beat.enabled">
                <line
                  :x1="beat.time * pixelsPerSecond"
                  y1="0"
                  :x2="beat.time * pixelsPerSecond"
                  :y2="trackHeight"
                  :stroke="beatFlash === beat.time ? '#4ade80' : 'rgba(74, 222, 128, 0.15)'"
                  stroke-width="1"
                />
              </g>
              
              <g 
                v-for="(kf, kfIdx) in track.keyframes" 
                :key="'kf'+kfIdx"
                class="keyframe-group"
                @mousedown.stop="onKeyframeMouseDown($event, trackIdx, kfIdx)"
                @dblclick="onKeyframeDblClick(trackIdx, kfIdx)"
              >
                <title>{{ formatTime(kf.time) }}{{ isKeyframeSynced(trackIdx, kfIdx) ? ' · 卡点成功' : '' }}</title>
                <polygon
                  :points="getKeyframePoints(kf.time)"
                  :fill="getKeyframeFill(trackIdx, kfIdx)"
                  :stroke="getKeyframeStroke(trackIdx, kfIdx)"
                  stroke-width="1.5"
                  class="keyframe-diamond"
                  :class="{ synced: isKeyframeSynced(trackIdx, kfIdx) }"
                />
                <circle
                  v-if="isKeyframeSynced(trackIdx, kfIdx)"
                  :cx="kf.time * pixelsPerSecond"
                  :cy="trackHeight / 2 - 14"
                  r="3"
                  fill="#4ade80"
                />
              </g>
            </svg>
          </div>

          <div 
            class="track-row audio-track-row"
            :class="{ 'has-audio': !!audioTrack }"
            :style="{ top: audioTrackTop + 'px', height: audioTrackHeight + 'px' }"
            @mousedown.stop="onAudioTrackMouseDown"
            @dblclick="onAudioTrackDblClick"
          >
            <svg class="audio-svg" :width="tracksWidth" :height="audioTrackHeight">
              <defs>
                <linearGradient id="audioFillGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#e8b059" stop-opacity="0.6" />
                  <stop offset="100%" stop-color="#e8b059" stop-opacity="0.2" />
                </linearGradient>
                <linearGradient id="audioFillProgress" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#4ade80" stop-opacity="0.7" />
                  <stop offset="100%" stop-color="#4ade80" stop-opacity="0.25" />
                </linearGradient>
              </defs>

              <rect x="0" y="0" :width="tracksWidth" :height="audioTrackHeight" fill="rgba(0,0,0,0.15)" />
              <line x1="0" :y1="audioTrackHeight / 2" :x2="tracksWidth" :y2="audioTrackHeight / 2" stroke="#4a4a7a" stroke-width="0.5" stroke-dasharray="2,3" />

              <g v-if="audioTrack && waveformData.length > 0">
                <polygon
                  :points="waveformPolygonPoints"
                  fill="url(#audioFillGrad)"
                />
                <polygon
                  :points="waveformProgressPolygonPoints"
                  fill="url(#audioFillProgress)"
                />
              </g>

              <g v-if="audioTrack" v-for="(beat, idx) in beats" :key="'ab'+idx">
                <rect
                  :x="beat.time * pixelsPerSecond - 2"
                  y="4"
                  width="4"
                  :height="audioTrackHeight - 8"
                  :rx="2"
                  :fill="beat.enabled ? (beatFlash === beat.time ? '#4ade80' : 'rgba(74, 222, 128, 0.55)') : 'rgba(160, 160, 192, 0.35)'"
                  class="beat-marker"
                  :class="{ enabled: beat.enabled, flash: beatFlash === beat.time }"
                  @mousedown.stop="onBeatMouseDown($event, idx)"
                />
                <text
                  v-if="beat.label"
                  :x="beat.time * pixelsPerSecond"
                  :y="audioTrackHeight - 6"
                  fill="#e0e0e0"
                  font-size="9"
                  text-anchor="middle"
                  pointer-events="none"
                >{{ beat.label }}</text>
                <line
                  :x1="beat.time * pixelsPerSecond"
                  y1="0"
                  :x2="beat.time * pixelsPerSecond"
                  :y2="audioTrackHeight"
                  :stroke="beat.enabled ? (beatFlash === beat.time ? '#4ade80' : 'rgba(74, 222, 128, 0.45)') : 'rgba(160, 160, 192, 0.25)'"
                  stroke-width="1"
                  pointer-events="none"
                />
              </g>

              <text v-if="!audioTrack" :x="tracksWidth / 2" :y="audioTrackHeight / 2 + 4" fill="#707090" font-size="12" text-anchor="middle" pointer-events="none">
                双击添加节拍 · 或上传鼓点音频
              </text>
            </svg>
          </div>

          <div 
            class="playhead"
            :style="{ left: playheadX + 'px' }"
          >
            <div class="playhead-line-vertical" :style="{ height: totalContentHeight + 'px' }"></div>
          </div>
        </div>
      </div>
    </div>

    <div class="timeline-footer">
      <div class="footer-left">
        <button class="footer-btn" @click="addKeyframe" :disabled="selectedTrackIndex < 0">
          <span>◆ 添加关键帧</span>
        </button>
        <button class="footer-btn" @click="deleteKeyframe" :disabled="!selectedKeyframe">
          <span>删除关键帧</span>
        </button>
        <span class="footer-divider"></span>
        <button class="footer-btn" @click="$emit('openAudioPanel')" :disabled="!script">
          <span>♪ 音频面板</span>
        </button>
        <button class="footer-btn success" @click="snapSelectedToBeat" :disabled="!selectedKeyframe || !enabledBeatCount">
          <span>✓ 吸附到最近节拍</span>
        </button>
      </div>
      <div class="footer-right">
        <span class="zoom-label">缩放:</span>
        <button class="zoom-btn" @click="zoomOut">−</button>
        <span class="zoom-value">{{ Math.round(zoomLevel * 100) }}%</span>
        <button class="zoom-btn" @click="zoomIn">+</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, onUnmounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useEditorStore } from '@/stores/editor'
import { formatTime } from '@/utils/animation'

const props = defineProps({
  trackHeight: { type: Number, default: 36 },
  rulerHeight: { type: Number, default: 40 },
  audioTrackHeight: { type: Number, default: 72 }
})

const emit = defineEmits(['timeChange', 'keyframeSelect', 'trackSelect', 'openAudioPanel'])

const store = useEditorStore()
const { currentTime, isPlaying, tracks, duration, audioTrack, beats, syncTolerance, beatFlash, script, audioElement } = storeToRefs(store)

const selectedTrackIndex = ref(0)
const selectedKeyframe = ref(null)
const zoomLevel = ref(1)
const rulerWidth = ref(600)
const rulerEl = ref(null)
const tracksEl = ref(null)
const bodyEl = ref(null)
const isDraggingPlayhead = ref(false)
const isDraggingKeyframe = ref(false)
const dragKeyframeData = ref(null)
const isDraggingBeat = ref(false)
const dragBeatIndex = ref(-1)

const waveformData = ref([])
const audioCtx = ref(null)

const pixelsPerSecond = computed(() => 100 * zoomLevel.value)
const tracksWidth = computed(() => Math.max(1, duration.value * pixelsPerSecond.value))
const tracksOnlyHeight = computed(() => tracks.value.length * props.trackHeight)
const audioTrackTop = computed(() => tracksOnlyHeight.value)
const totalContentHeight = computed(() => tracksOnlyHeight.value + (props.audioTrackHeight + 4))

const playheadX = computed(() => currentTime.value * pixelsPerSecond.value)

const enabledBeatCount = computed(() => beats.value.filter(b => b.enabled).length)

const syncedCount = computed(() => {
  let n = 0
  for (let ti = 0; ti < tracks.value.length; ti++) {
    const track = tracks.value[ti]
    if (!track?.keyframes) continue
    for (let ki = 0; ki < track.keyframes.length; ki++) {
      if (store.isKeyframeSynced(ti, ki)) n++
    }
  }
  return n
})

const isAudioPlaying = computed(() => store.isAudioPlaying)

const ticks = computed(() => {
  const result = []
  const totalDuration = duration.value
  let interval = 1
  if (zoomLevel.value >= 2) interval = 0.5
  if (zoomLevel.value >= 4) interval = 0.25
  if (zoomLevel.value <= 0.5) interval = 2
  if (zoomLevel.value <= 0.25) interval = 5
  for (let t = 0; t <= totalDuration + interval; t += interval) {
    result.push({
      time: t,
      x: t * pixelsPerSecond.value,
      major: t % (interval * 2) < 0.001 || interval < 1
    })
  }
  return result
})

const waveformPolygonPoints = computed(() => {
  if (waveformData.value.length === 0) return ''
  const data = waveformData.value
  const mid = props.audioTrackHeight / 2
  const amp = (props.audioTrackHeight / 2) * 0.85
  const pts = []
  pts.push(`0,${mid}`)
  data.forEach((v, i) => {
    const x = (i / (data.length - 1)) * tracksWidth.value
    pts.push(`${x},${mid - v * amp}`)
  })
  for (let i = data.length - 1; i >= 0; i--) {
    const x = (i / (data.length - 1)) * tracksWidth.value
    pts.push(`${x},${mid + data[i] * amp}`)
  }
  return pts.join(' ')
})

const waveformProgressPolygonPoints = computed(() => {
  if (waveformData.value.length === 0) return ''
  const data = waveformData.value
  const progressRatio = Math.min(1, currentTime.value / duration.value)
  const progressIdx = Math.floor(progressRatio * (data.length - 1))
  if (progressIdx <= 0) return ''
  const mid = props.audioTrackHeight / 2
  const amp = (props.audioTrackHeight / 2) * 0.85
  const pts = []
  pts.push(`0,${mid}`)
  for (let i = 0; i <= progressIdx; i++) {
    const x = (i / (data.length - 1)) * tracksWidth.value
    pts.push(`${x},${mid - data[i] * amp}`)
  }
  for (let i = progressIdx; i >= 0; i--) {
    const x = (i / (data.length - 1)) * tracksWidth.value
    pts.push(`${x},${mid + data[i] * amp}`)
  }
  return pts.join(' ')
})

function formatTickTime(time) {
  if (time >= 60) {
    const mins = Math.floor(time / 60)
    const secs = time % 60
    return `${mins}:${secs.toFixed(secs % 1 === 0 ? 0 : 1).padStart(2, '0')}`
  }
  return time.toFixed(time % 1 === 0 ? 0 : 1) + 's'
}

function getKeyframePoints(time) {
  const x = time * pixelsPerSecond.value
  const y = props.trackHeight / 2
  const size = 8
  return `${x},${y - size} ${x + size},${y} ${x},${y + size} ${x - size},${y}`
}

function getKeyframeFill(trackIdx, kfIdx) {
  const isSelected = selectedKeyframe.value?.trackIdx === trackIdx && selectedKeyframe.value?.kfIdx === kfIdx
  if (store.isKeyframeSynced(trackIdx, kfIdx)) {
    return isSelected ? '#86efac' : '#4ade80'
  }
  return isSelected ? '#e8b059' : '#8a8ab0'
}

function getKeyframeStroke(trackIdx, kfIdx) {
  if (store.isKeyframeSynced(trackIdx, kfIdx)) return '#166534'
  return '#ffffff'
}

function isKeyframeSynced(trackIdx, kfIdx) {
  return store.isKeyframeSynced(trackIdx, kfIdx)
}

function togglePlay() {
  store.togglePlay()
}

function stop() {
  store.stop()
}

function selectTrack(index) {
  selectedTrackIndex.value = index
  selectedKeyframe.value = null
  store.selectBone(tracks.value[index].boneId)
  emit('trackSelect', index)
}

function onRulerMouseDown(event) {
  if (event.target.closest('.playhead')) return
  const rect = rulerEl.value.getBoundingClientRect()
  const x = event.clientX - rect.left
  const time = x / pixelsPerSecond.value
  store.setTime(Math.max(0, Math.min(duration.value, time)))
  isDraggingPlayhead.value = true
  emit('timeChange', currentTime.value)
}

function onRulerMouseMove(event) {
  if (!isDraggingPlayhead.value) return
  const rect = rulerEl.value.getBoundingClientRect()
  const x = event.clientX - rect.left
  const time = x / pixelsPerSecond.value
  store.setTime(Math.max(0, Math.min(duration.value, time)))
  emit('timeChange', currentTime.value)
}

function onRulerMouseUp() {
  isDraggingPlayhead.value = false
}

function onPlayheadMouseDown(event) {
  event.stopPropagation()
  isDraggingPlayhead.value = true
}

function onTracksMouseDown(event) {
  if (event.target.closest('.keyframe-group')) return
  if (event.target.closest('.beat-marker')) return
  if (event.target.closest('.audio-track-row')) return
  const rect = tracksEl.value.getBoundingClientRect()
  const x = event.clientX - rect.left + tracksEl.value.scrollLeft
  const time = x / pixelsPerSecond.value
  store.setTime(Math.max(0, Math.min(duration.value, time)))
  isDraggingPlayhead.value = true
  selectedKeyframe.value = null
  emit('timeChange', currentTime.value)
}

function onTracksMouseMove(event) {
  if (isDraggingPlayhead.value) {
    const rect = tracksEl.value.getBoundingClientRect()
    const x = event.clientX - rect.left + tracksEl.value.scrollLeft
    const time = x / pixelsPerSecond.value
    store.setTime(Math.max(0, Math.min(duration.value, time)))
    emit('timeChange', currentTime.value)
  }
  if (isDraggingKeyframe.value && dragKeyframeData.value) {
    const rect = tracksEl.value.getBoundingClientRect()
    const x = event.clientX - rect.left + tracksEl.value.scrollLeft
    const newTime = x / pixelsPerSecond.value
    const snapTime = Math.round(newTime * 10) / 10
    store.moveKeyframe(dragKeyframeData.value.trackIdx, dragKeyframeData.value.kfIdx, snapTime)
  }
  if (isDraggingBeat.value && dragBeatIndex.value >= 0) {
    const rect = tracksEl.value.getBoundingClientRect()
    const x = event.clientX - rect.left + tracksEl.value.scrollLeft
    let newTime = x / pixelsPerSecond.value
    newTime = Math.max(0, Math.min(duration.value, Math.round(newTime * 100) / 100))
    if (script.value?.audioTrack?.beats[dragBeatIndex.value]) {
      script.value.audioTrack.beats[dragBeatIndex.value].time = newTime
      store.checkAllKeyframesSynced()
    }
  }
}

function onTracksMouseUp() {
  isDraggingPlayhead.value = false
  isDraggingKeyframe.value = false
  dragKeyframeData.value = null
  if (isDraggingBeat.value && dragBeatIndex.value >= 0) {
    isDraggingBeat.value = false
    dragBeatIndex.value = -1
    if (script.value?.audioTrack?.beats) {
      store.saveBeats(script.value.audioTrack.beats).catch(() => {})
    }
  }
}

function onKeyframeMouseDown(event, trackIdx, kfIdx) {
  selectTrack(trackIdx)
  selectedKeyframe.value = { trackIdx, kfIdx }
  isDraggingKeyframe.value = true
  dragKeyframeData.value = { trackIdx, kfIdx }
  emit('keyframeSelect', { trackIdx, kfIdx })
}

function onKeyframeDblClick(trackIdx, kfIdx) {
  if (!enabledBeatCount.value) return
  const track = tracks.value[trackIdx]
  if (!track) return
  const kf = track.keyframes[kfIdx]
  if (!kf) return
  const tol = syncTolerance.value
  let nearest = -1
  let minD = Infinity
  beats.value.forEach((b, bi) => {
    if (!b.enabled) return
    const d = Math.abs(kf.time - b.time)
    if (d < minD) { minD = d; nearest = bi }
  })
  if (nearest >= 0 && minD <= tol * 5) {
    store.moveKeyframe(trackIdx, kfIdx, beats.value[nearest].time)
  }
}

function addKeyframe() {
  if (selectedTrackIndex.value < 0) return
  const track = tracks.value[selectedTrackIndex.value]
  const bone = store.bones.find(b => b.id === track.boneId)
  const angle = bone ? store.getBoneAngle(track.boneId) : 0
  store.addKeyframe(selectedTrackIndex.value, currentTime.value, angle)
}

function deleteKeyframe() {
  if (!selectedKeyframe.value) return
  store.deleteKeyframe(selectedKeyframe.value.trackIdx, selectedKeyframe.value.kfIdx)
  selectedKeyframe.value = null
}

function snapSelectedToBeat() {
  if (!selectedKeyframe.value || !enabledBeatCount.value) return
  const { trackIdx, kfIdx } = selectedKeyframe.value
  const track = tracks.value[trackIdx]
  if (!track) return
  const kf = track.keyframes[kfIdx]
  if (!kf) return
  let nearest = -1
  let minD = Infinity
  beats.value.forEach((b, bi) => {
    if (!b.enabled) return
    const d = Math.abs(kf.time - b.time)
    if (d < minD) { minD = d; nearest = bi }
  })
  if (nearest >= 0) {
    store.moveKeyframe(trackIdx, kfIdx, beats.value[nearest].time)
  }
}

function zoomIn() {
  zoomLevel.value = Math.min(zoomLevel.value * 1.5, 8)
}
function zoomOut() {
  zoomLevel.value = Math.max(zoomLevel.value / 1.5, 0.125)
}

function onTracksWheel(event) {
  if (event.ctrlKey || event.metaKey) {
    event.preventDefault()
    if (event.deltaY < 0) {
      zoomIn()
    } else {
      zoomOut()
    }
  }
}

function onAudioTrackMouseDown(event) {
  const rect = tracksEl.value.getBoundingClientRect()
  const x = event.clientX - rect.left + tracksEl.value.scrollLeft
  const time = x / pixelsPerSecond.value
  store.setTime(Math.max(0, Math.min(duration.value, time)))
  isDraggingPlayhead.value = true
  emit('timeChange', currentTime.value)
}

function onAudioTrackDblClick(event) {
  const rect = tracksEl.value.getBoundingClientRect()
  const x = event.clientX - rect.left + tracksEl.value.scrollLeft
  const time = Math.max(0, Math.min(duration.value, Math.round((x / pixelsPerSecond.value) * 100) / 100))
  if (!script.value?.audioTrack) {
    if (!script.value) return
    script.value.audioTrack = {
      fileName: '',
      fileSize: 0,
      duration: duration.value,
      volume: 1,
      startTime: 0,
      beats: [],
      syncTolerance: 0.08
    }
  }
  store.addBeat(time)
  store.saveBeats(script.value.audioTrack.beats).catch(() => {})
}

function onBeatMouseDown(event, idx) {
  event.preventDefault()
  isDraggingBeat.value = true
  dragBeatIndex.value = idx
}

async function buildWaveform() {
  if (!audioTrack.value?.fileName || !store.audioUrl) {
    waveformData.value = []
    return
  }
  try {
    if (!audioCtx.value) {
      const AC = window.AudioContext || window.webkitAudioContext
      if (!AC) return
      audioCtx.value = new AC()
    }
    const resp = await fetch(store.audioUrl)
    const buf = await resp.arrayBuffer()
    let decoded
    try {
      decoded = await audioCtx.value.decodeAudioData(buf)
    } catch {
      return
    }
    const samples = 400
    const raw = decoded.getChannelData(0)
    const blockSize = Math.floor(raw.length / samples)
    const result = new Array(samples).fill(0)
    for (let i = 0; i < samples; i++) {
      let sum = 0
      const start = i * blockSize
      for (let j = 0; j < blockSize; j++) {
        sum += Math.abs(raw[start + j] || 0)
      }
      result[i] = Math.min(1, (sum / blockSize) * 3)
    }
    waveformData.value = result
  } catch (e) {
    console.warn('waveform build failed', e)
  }
}

watch(
  () => [audioTrack.value?.fileName, store.audioUrl, duration.value],
  () => {
    buildWaveform()
  },
  { immediate: true }
)

onMounted(() => {
  if (rulerEl.value) {
    rulerWidth.value = rulerEl.value.clientWidth
  }
  if (tracks.value.length > 0) {
    store.selectBone(tracks.value[0].boneId)
  }
})

onUnmounted(() => {
  if (audioCtx.value) {
    audioCtx.value.close().catch(() => {})
  }
})
</script>

<style scoped>
.timeline-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1e1e38;
  border-radius: 8px;
  overflow: hidden;
}

.timeline-header {
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  border-bottom: 1px solid #3a3a5a;
}

.timeline-controls {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 10px 16px;
  background: #252545;
}

.play-controls {
  display: flex;
  gap: 6px;
}

.play-btn {
  width: 36px;
  height: 36px;
  border-radius: 6px;
  background: #3a3a5a;
  color: #e0e0e0;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.play-btn:hover {
  background: #4a4a7a;
}

.play-btn.play-main {
  background: linear-gradient(135deg, #e8b059, #d4944a);
  color: #1a1a2e;
}

.play-btn.play-main:hover {
  background: linear-gradient(135deg, #f4c069, #e8a45a);
}

.btn-icon {
  font-size: 14px;
  line-height: 1;
}

.time-display {
  display: flex;
  align-items: baseline;
  gap: 4px;
  font-family: 'Courier New', monospace;
}

.time-current {
  font-size: 16px;
  font-weight: 600;
  color: #e8b059;
}

.time-sep {
  color: #6a6a9a;
}

.time-total {
  font-size: 12px;
  color: #8080a0;
}

.audio-status {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 12px;
  margin-left: auto;
  background: rgba(74, 222, 128, 0.08);
  border: 1px solid rgba(74, 222, 128, 0.2);
  border-radius: 6px;
}

.audio-status-icon {
  font-size: 16px;
  color: #4ade80;
}

.audio-status-icon.muted {
  color: #6a6a9a;
  opacity: 0.5;
}

.audio-status-icon.active {
  animation: pulse 0.8s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.2); opacity: 0.7; }
}

.audio-status-text {
  display: flex;
  flex-direction: column;
}

.audio-filename {
  font-size: 11px;
  color: #c0c0e0;
  font-weight: 500;
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.muted-text {
  font-size: 11px;
  color: #707090;
}

.audio-beats {
  font-size: 10px;
  color: #8080a0;
}

.audio-sync {
  padding-left: 10px;
  border-left: 1px solid rgba(128, 128, 160, 0.3);
}

.sync-count {
  font-size: 12px;
  font-weight: 600;
}

.timeline-ruler-container {
  display: flex;
  border-top: 1px solid #3a3a5a;
}

.ruler-track-label {
  width: 140px;
  flex-shrink: 0;
  padding: 0 12px;
  display: flex;
  align-items: center;
  background: #252545;
  border-right: 1px solid #3a3a5a;
  font-size: 11px;
  color: #8080a0;
  font-weight: 500;
}

.timeline-ruler {
  flex: 1;
  position: relative;
  height: 40px;
  background: #2a2a4a;
  overflow: hidden;
  cursor: pointer;
}

.ruler-svg {
  display: block;
  width: 100%;
  height: 100%;
}

.playhead {
  position: absolute;
  top: 0;
  pointer-events: auto;
  z-index: 10;
}

.playhead-triangle {
  width: 0;
  height: 0;
  border-left: 6px solid transparent;
  border-right: 6px solid transparent;
  border-top: 8px solid #e8b059;
  transform: translateX(-50%);
  cursor: ew-resize;
}

.playhead-line {
  position: absolute;
  top: 8px;
  left: -1px;
  width: 2px;
  height: 32px;
  background: #e8b059;
  pointer-events: none;
}

.timeline-body {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.track-labels {
  width: 140px;
  flex-shrink: 0;
  overflow-y: auto;
  background: #252545;
  border-right: 1px solid #3a3a5a;
}

.track-label {
  height: 36px;
  padding: 0 12px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  border-bottom: 1px solid #3a3a5a;
  cursor: pointer;
  transition: background 0.15s;
}

.track-label:hover {
  background: #3a3a5a;
}

.track-label.selected {
  background: rgba(232, 176, 89, 0.15);
}

.track-label.selected .track-name {
  color: #e8b059;
}

.track-label.audio-label {
  height: 72px;
  border-top: 2px solid rgba(74, 222, 128, 0.2);
}

.track-label.audio-label.has-audio {
  background: rgba(74, 222, 128, 0.08);
}

.track-label.audio-label.has-audio .track-name {
  color: #4ade80;
}

.audio-icon {
  margin-right: 4px;
}

.track-name {
  font-size: 12px;
  font-weight: 500;
  color: #c0c0e0;
}

.track-bone-id {
  font-size: 10px;
  color: #707090;
}

.tracks-container {
  flex: 1;
  overflow: auto;
  position: relative;
  background: #1a1a2e;
}

.tracks-content {
  position: relative;
}

.track-row {
  position: absolute;
  left: 0;
  right: 0;
  height: 36px;
  cursor: pointer;
}

.track-row.selected {
  background: rgba(232, 176, 89, 0.08);
}

.track-svg {
  display: block;
}

.keyframe-group {
  cursor: grab;
}

.keyframe-group:active {
  cursor: grabbing;
}

.keyframe-diamond {
  transition: fill 0.15s, stroke 0.15s;
  filter: drop-shadow(0 1px 1px rgba(0,0,0,0.3));
}

.keyframe-diamond.synced {
  filter: drop-shadow(0 0 6px rgba(74, 222, 128, 0.6));
}

.keyframe-group:hover .keyframe-diamond {
  opacity: 0.9;
}

.audio-track-row {
  position: absolute;
  left: 0;
  right: 0;
  border-top: 2px solid rgba(74, 222, 128, 0.25);
  background: rgba(74, 222, 128, 0.03);
  cursor: pointer;
}

.audio-track-row.has-audio {
  background: rgba(74, 222, 128, 0.06);
}

.audio-svg {
  display: block;
}

.beat-marker {
  cursor: grab;
  transition: fill 0.15s;
}

.beat-marker:hover {
  fill: rgba(74, 222, 128, 0.85) !important;
}

.beat-marker:active {
  cursor: grabbing;
}

.beat-marker.flash {
  animation: beatFlash 0.25s ease-out;
}

@keyframes beatFlash {
  0% { filter: brightness(2); transform: scaleY(1.4); transform-origin: center; }
  100% { filter: brightness(1); transform: scaleY(1); }
}

.playhead-line-vertical {
  position: absolute;
  top: 0;
  left: -1px;
  width: 2px;
  background: rgba(232, 176, 89, 0.6);
  pointer-events: none;
}

.timeline-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background: #252545;
  border-top: 1px solid #3a3a5a;
  flex-shrink: 0;
}

.footer-left,
.footer-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.footer-divider {
  width: 1px;
  height: 20px;
  background: #3a3a5a;
  margin: 0 4px;
}

.footer-btn {
  padding: 6px 12px;
  border-radius: 4px;
  background: #3a3a5a;
  color: #c0c0e0;
  font-size: 12px;
  transition: all 0.2s;
}

.footer-btn:hover:not(:disabled) {
  background: #4a4a7a;
  color: #e0e0e0;
}

.footer-btn.success {
  background: rgba(74, 222, 128, 0.2);
  color: #4ade80;
  border: 1px solid rgba(74, 222, 128, 0.4);
}

.footer-btn.success:hover:not(:disabled) {
  background: rgba(74, 222, 128, 0.35);
  color: #86efac;
}

.footer-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.zoom-label {
  font-size: 12px;
  color: #8080a0;
}

.zoom-btn {
  width: 24px;
  height: 24px;
  border-radius: 4px;
  background: #3a3a5a;
  color: #c0c0e0;
  font-size: 14px;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.zoom-btn:hover {
  background: #4a4a7a;
}

.zoom-value {
  font-size: 12px;
  color: #a0a0c0;
  min-width: 48px;
  text-align: center;
}
</style>
