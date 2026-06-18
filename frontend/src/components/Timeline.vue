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
          <svg class="ruler-svg" :width="rulerWidth" :height="rulerHeight">
            <g v-for="tick in ticks" :key="tick.time">
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
          :style="{ width: tracksWidth + 'px', height: tracksHeight + 'px' }"
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
              
              <g 
                v-for="(kf, kfIdx) in track.keyframes" 
                :key="kfIdx"
                class="keyframe-group"
                @mousedown.stop="onKeyframeMouseDown($event, trackIdx, kfIdx)"
              >
                <polygon
                  :points="getKeyframePoints(kf.time)"
                  :fill="selectedKeyframe?.trackIdx === trackIdx && selectedKeyframe?.kfIdx === kfIdx ? '#e8b059' : '#8a8ab0'"
                  stroke="#fff"
                  stroke-width="1"
                  class="keyframe-diamond"
                />
              </g>
            </svg>
          </div>

          <div 
            class="playhead"
            :style="{ left: playheadX + 'px' }"
          >
            <div class="playhead-line-vertical"></div>
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

<script setup>import { ref, computed, onMounted } from 'vue';
import { storeToRefs } from 'pinia';
import { useEditorStore } from '@/stores/editor';
import { formatTime } from '@/utils/animation';
const props = defineProps({
 trackHeight: { type: Number, default: 36 },
 rulerHeight: { type: Number, default: 40 }
});
const emit = defineEmits(['timeChange', 'keyframeSelect', 'trackSelect']);
const store = useEditorStore();
const { currentTime, isPlaying, tracks, duration } = storeToRefs(store);
const selectedTrackIndex = ref(0);
const selectedKeyframe = ref(null);
const zoomLevel = ref(1);
const rulerWidth = ref(600);
const rulerEl = ref(null);
const tracksEl = ref(null);
const bodyEl = ref(null);
const isDraggingPlayhead = ref(false);
const isDraggingKeyframe = ref(false);
const dragKeyframeData = ref(null);
const pixelsPerSecond = computed(() => 100 * zoomLevel.value);
const tracksWidth = computed(() => duration.value * pixelsPerSecond.value);
const tracksHeight = computed(() => tracks.value.length * props.trackHeight);
const playheadX = computed(() => currentTime.value * pixelsPerSecond.value);
const ticks = computed(() => {
 const result = [];
 const totalDuration = duration.value;
 let interval = 1;
 if (zoomLevel.value >= 2)
 interval = 0.5;
 if (zoomLevel.value >= 4)
 interval = 0.25;
 if (zoomLevel.value <= 0.5)
 interval = 2;
 if (zoomLevel.value <= 0.25)
 interval = 5;
 for (let t = 0; t <= totalDuration + interval; t += interval) {
 result.push({
 time: t,
 x: t * pixelsPerSecond.value,
 major: t % (interval * 2) < 0.001 || interval < 1
 });
 }
 return result;
});
function formatTickTime(time) {
 if (time >= 60) {
 const mins = Math.floor(time / 60);
 const secs = time % 60;
 return `${mins}:${secs.toFixed(secs % 1 === 0 ? 0 : 1).padStart(2, '0')}`;
 }
 return time.toFixed(time % 1 === 0 ? 0 : 1) + 's';
}
function getKeyframePoints(time) {
 const x = time * pixelsPerSecond.value;
 const y = props.trackHeight / 2;
 const size = 8;
 return `${x},${y - size} ${x + size},${y} ${x},${y + size} ${x - size},${y}`;
}
function togglePlay() {
 store.togglePlay();
}
function stop() {
 store.stop();
}
function selectTrack(index) {
 selectedTrackIndex.value = index;
 selectedKeyframe.value = null;
 store.selectBone(tracks.value[index].boneId);
 emit('trackSelect', index);
}
function onRulerMouseDown(event) {
 if (event.target.closest('.playhead'))
 return;
 const rect = rulerEl.value.getBoundingClientRect();
 const x = event.clientX - rect.left;
 const time = x / pixelsPerSecond.value;
 store.setTime(Math.max(0, Math.min(duration.value, time)));
 isDraggingPlayhead.value = true;
 emit('timeChange', currentTime.value);
}
function onRulerMouseMove(event) {
 if (!isDraggingPlayhead.value)
 return;
 const rect = rulerEl.value.getBoundingClientRect();
 const x = event.clientX - rect.left;
 const time = x / pixelsPerSecond.value;
 store.setTime(Math.max(0, Math.min(duration.value, time)));
 emit('timeChange', currentTime.value);
}
function onRulerMouseUp() {
 isDraggingPlayhead.value = false;
}
function onPlayheadMouseDown(event) {
 event.stopPropagation();
 isDraggingPlayhead.value = true;
}
function onTracksMouseDown(event) {
 if (event.target.closest('.keyframe-group'))
 return;
 const rect = tracksEl.value.getBoundingClientRect();
 const x = event.clientX - rect.left + tracksEl.value.scrollLeft;
 const time = x / pixelsPerSecond.value;
 store.setTime(Math.max(0, Math.min(duration.value, time)));
 isDraggingPlayhead.value = true;
 selectedKeyframe.value = null;
 emit('timeChange', currentTime.value);
}
function onTracksMouseMove(event) {
 if (isDraggingPlayhead.value) {
 const rect = tracksEl.value.getBoundingClientRect();
 const x = event.clientX - rect.left + tracksEl.value.scrollLeft;
 const time = x / pixelsPerSecond.value;
 store.setTime(Math.max(0, Math.min(duration.value, time)));
 emit('timeChange', currentTime.value);
 }
 if (isDraggingKeyframe.value && dragKeyframeData.value) {
 const rect = tracksEl.value.getBoundingClientRect();
 const x = event.clientX - rect.left + tracksEl.value.scrollLeft;
 const newTime = x / pixelsPerSecond.value;
 const snapTime = Math.round(newTime * 10) / 10;
 store.moveKeyframe(dragKeyframeData.value.trackIdx, dragKeyframeData.value.kfIdx, snapTime);
 }
}
function onTracksMouseUp() {
 isDraggingPlayhead.value = false;
 isDraggingKeyframe.value = false;
 dragKeyframeData.value = null;
}
function onKeyframeMouseDown(event, trackIdx, kfIdx) {
 selectTrack(trackIdx);
 selectedKeyframe.value = { trackIdx, kfIdx };
 isDraggingKeyframe.value = true;
 dragKeyframeData.value = { trackIdx, kfIdx };
 emit('keyframeSelect', { trackIdx, kfIdx });
}
function addKeyframe() {
 if (selectedTrackIndex.value < 0)
 return;
 const track = tracks.value[selectedTrackIndex.value];
 const bone = store.bones.find(b => b.id === track.boneId);
 const angle = bone ? store.getBoneAngle(track.boneId) : 0;
 store.addKeyframe(selectedTrackIndex.value, currentTime.value, angle);
}
function deleteKeyframe() {
 if (!selectedKeyframe.value)
 return;
 store.deleteKeyframe(selectedKeyframe.value.trackIdx, selectedKeyframe.value.kfIdx);
 selectedKeyframe.value = null;
}
function zoomIn() {
 zoomLevel.value = Math.min(zoomLevel.value * 1.5, 8);
}
function zoomOut() {
 zoomLevel.value = Math.max(zoomLevel.value / 1.5, 0.125);
}
function onTracksWheel(event) {
 if (event.ctrlKey || event.metaKey) {
 event.preventDefault();
 if (event.deltaY < 0) {
 zoomIn();
 }
 else {
 zoomOut();
 }
 }
}
onMounted(() => {
 if (rulerEl.value) {
 rulerWidth.value = rulerEl.value.clientWidth;
 }
 if (tracks.value.length > 0) {
 store.selectBone(tracks.value[0].boneId);
 }
});
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
  transition: fill 0.15s;
}

.keyframe-group:hover .keyframe-diamond {
  fill: #b0b0d0;
}

.playhead-line-vertical {
  position: absolute;
  top: 0;
  left: -1px;
  width: 2px;
  height: 100vh;
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
