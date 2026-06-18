<template>
  <div class="puppet-stage">
    <div class="stage-header">
      <span class="stage-title">{{ character?.name || '皮影人物' }}</span>
      <span class="stage-subtitle">{{ character?.category || '' }}</span>
    </div>
    <div class="stage-canvas">
      <svg 
        ref="svgEl"
        :viewBox="`0 0 ${viewWidth} ${viewHeight}`"
        class="puppet-svg"
        @mousedown="onSvgMouseDown"
        @mousemove="onSvgMouseMove"
        @mouseup="onSvgMouseUp"
        @mouseleave="onSvgMouseUp"
      >
        <defs>
          <radialGradient id="stageGlow" cx="50%" cy="50%" r="50%">
            <stop offset="0%" style="stop-color:#4a3a6a;stop-opacity:0.3" />
            <stop offset="100%" style="stop-color:#1a1a2e;stop-opacity:0" />
          </radialGradient>
          <filter id="shadow" x="-20%" y="-20%" width="140%" height="140%">
            <feDropShadow dx="3" dy="3" stdDeviation="3" flood-color="#000" flood-opacity="0.4" />
          </filter>
        </defs>

        <rect :width="viewWidth" :height="viewHeight" fill="url(#stageGlow)" />
        
        <line 
          :x1="0" 
          :y1="viewHeight * 0.85" 
          :x2="viewWidth" 
          :y2="viewHeight * 0.85" 
          stroke="#3a3a5a" 
          stroke-width="1"
          stroke-dasharray="5,5"
        />

        <g :transform="`scale(${scale})`" filter="url(#shadow)">
          <g 
            v-for="bone in visibleBones" 
            :key="bone.id"
            class="bone-group"
            :class="{ selected: selectedBoneId === bone.id, interactive: bone.id !== 'root' }"
            @mousedown.stop="onBoneMouseDown($event, bone.id)"
          >
            <line
              v-if="bone.id !== 'root'"
              :x1="bone.startX"
              :y1="bone.startY"
              :x2="bone.endX"
              :y2="bone.endY"
              :stroke="bone.color"
              :stroke-width="bone.width"
              stroke-linecap="round"
              class="bone-line"
            />
            
            <circle
              v-if="bone.id === 'head'"
              :cx="bone.endX"
              :cy="bone.endY"
              :r="bone.width / 2"
              :fill="bone.color"
              class="bone-head"
            />
            
            <circle
              v-if="bone.id !== 'root'"
              :cx="bone.startX"
              :cy="bone.startY"
              r="5"
              fill="#fff"
              stroke="#6a6a9a"
              stroke-width="2"
              class="joint-circle"
            />
            
            <circle
              v-if="bone.id !== 'root' && selectedBoneId === bone.id"
              :cx="bone.endX"
              :cy="bone.endY"
              r="8"
              fill="#e8b059"
              stroke="#fff"
              stroke-width="2"
              class="end-handle"
            />
          </g>
        </g>
      </svg>
    </div>
    
    <div v-if="selectedBone" class="bone-controls">
      <div class="control-title">
        <span class="control-bone-name">{{ selectedBone.name }}</span>
        <span class="control-bone-id">{{ selectedBone.id }}</span>
      </div>
      <div class="control-slider">
        <label>角度: {{ currentAngle.toFixed(1) }}°</label>
        <input 
          type="range" 
          :min="selectedBone.minAngle" 
          :max="selectedBone.maxAngle" 
          :value="currentAngle"
          @input="onAngleChange"
          class="angle-slider"
        />
      </div>
      <div class="control-buttons">
        <button @click="resetBoneAngle" class="btn-reset">重置</button>
        <button @click="addKeyframeAtCurrent" class="btn-keyframe">添加关键帧</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useEditorStore } from '@/stores/editor'
import { degToRad } from '@/utils/animation'

const props = defineProps({
  viewWidth: { type: Number, default: 600 },
  viewHeight: { type: Number, default: 500 },
  scale: { type: Number, default: 1 }
})

const emit = defineEmits(['boneSelect', 'angleChange'])

const store = useEditorStore()
const { bonePositions, selectedBoneId, currentTime } = storeToRefs(store)

const character = computed(() => store.character)
const visibleBones = computed(() => Object.values(bonePositions.value))

const selectedBone = computed(() => {
  if (!selectedBoneId.value) return null
  return bonePositions.value[selectedBoneId.value]
})

const currentAngle = computed(() => {
  if (!selectedBoneId.value) return 0
  const pos = bonePositions.value[selectedBoneId.value]
  return pos ? pos.currentAngle : 0
})

const isDragging = ref(false)
const dragBoneId = ref(null)
const svgEl = ref(null)

function onBoneMouseDown(event, boneId) {
  if (boneId === 'root') return
  store.selectBone(boneId)
  isDragging.value = true
  dragBoneId.value = boneId
  emit('boneSelect', boneId)
}

function onSvgMouseDown() {
}

function getSvgPoint(event) {
  const svg = svgEl.value
  if (!svg) return { x: 0, y: 0 }
  
  const rect = svg.getBoundingClientRect()
  const x = (event.clientX - rect.left) / rect.width * props.viewWidth / props.scale
  const y = (event.clientY - rect.top) / rect.height * props.viewHeight / props.scale
  return { x, y }
}

function onSvgMouseMove(event) {
  if (!isDragging.value || !dragBoneId.value) return
  
  const bone = bonePositions.value[dragBoneId.value]
  if (!bone) return
  
  const point = getSvgPoint(event)
  const dx = point.x - bone.startX
  const dy = point.y - bone.startY
  let angle = Math.atan2(dy, dx) * 180 / Math.PI
  
  angle = Math.max(bone.minAngle, Math.min(bone.maxAngle, angle))
  
  store.setBoneAngle(dragBoneId.value, angle)
  emit('angleChange', { boneId: dragBoneId.value, angle })
}

function onSvgMouseUp() {
  isDragging.value = false
  dragBoneId.value = null
}

function onAngleChange(event) {
  const angle = parseFloat(event.target.value)
  if (selectedBoneId.value) {
    store.setBoneAngle(selectedBoneId.value, angle)
    emit('angleChange', { boneId: selectedBoneId.value, angle })
  }
}

function resetBoneAngle() {
  if (selectedBoneId.value && character.value) {
    const bone = character.value.bones.find(b => b.id === selectedBoneId.value)
    if (bone) {
      store.setBoneAngle(selectedBoneId.value, bone.baseAngle)
    }
  }
}

function addKeyframeAtCurrent() {
  if (selectedBoneId.value) {
    const trackIdx = store.tracks.findIndex(t => t.boneId === selectedBoneId.value)
    if (trackIdx >= 0) {
      store.addKeyframe(trackIdx, currentTime.value, currentAngle.value)
    }
  }
}
</script>

<style scoped>
.puppet-stage {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #252545;
  border-radius: 8px;
  overflow: hidden;
}

.stage-header {
  display: flex;
  align-items: baseline;
  gap: 12px;
  padding: 12px 16px;
  background: linear-gradient(135deg, #2d1b4e 0%, #252545 100%);
  border-bottom: 1px solid #3d2d5e;
}

.stage-title {
  font-size: 16px;
  font-weight: 600;
  color: #e8b059;
}

.stage-subtitle {
  font-size: 12px;
  color: #8080a0;
}

.stage-canvas {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  background: radial-gradient(ellipse at center, #2a2a4a 0%, #1e1e38 100%);
}

.puppet-svg {
  width: 100%;
  height: 100%;
  max-width: 100%;
  max-height: 100%;
  cursor: default;
}

.bone-group {
  cursor: pointer;
  transition: opacity 0.2s;
}

.bone-group.interactive:hover .bone-line,
.bone-group.interactive:hover .bone-head {
  filter: brightness(1.2);
}

.bone-group.selected .bone-line,
.bone-group.selected .bone-head {
  filter: drop-shadow(0 0 6px rgba(232, 176, 89, 0.6));
}

.bone-line {
  transition: stroke 0.2s;
}

.joint-circle {
  pointer-events: none;
}

.end-handle {
  cursor: grab;
}

.bone-controls {
  padding: 12px 16px;
  background: #1e1e38;
  border-top: 1px solid #3d2d5e;
}

.control-title {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 10px;
}

.control-bone-name {
  font-size: 14px;
  font-weight: 600;
  color: #e0e0e0;
}

.control-bone-id {
  font-size: 11px;
  color: #707090;
}

.control-slider {
  margin-bottom: 10px;
}

.control-slider label {
  display: block;
  font-size: 12px;
  color: #a0a0c0;
  margin-bottom: 6px;
}

.angle-slider {
  width: 100%;
  height: 6px;
  -webkit-appearance: none;
  appearance: none;
  background: #3a3a5a;
  border-radius: 3px;
  outline: none;
}

.angle-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 16px;
  height: 16px;
  background: #e8b059;
  border-radius: 50%;
  cursor: pointer;
  transition: transform 0.15s;
}

.angle-slider::-webkit-slider-thumb:hover {
  transform: scale(1.2);
}

.control-buttons {
  display: flex;
  gap: 8px;
}

.control-buttons button {
  flex: 1;
  padding: 8px 12px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  transition: all 0.2s;
}

.btn-reset {
  background: #3a3a5a;
  color: #b0b0d0;
}

.btn-reset:hover {
  background: #4a4a7a;
  color: #e0e0e0;
}

.btn-keyframe {
  background: linear-gradient(135deg, #e8b059, #d4944a);
  color: #1a1a2e;
}

.btn-keyframe:hover {
  background: linear-gradient(135deg, #f4c069, #e8a45a);
}
</style>
