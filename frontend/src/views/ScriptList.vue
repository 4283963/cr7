<template>
  <div class="script-list-page">
    <div class="page-header">
      <div class="page-title-section">
        <h1 class="page-title">动作剧本管理</h1>
        <p class="page-subtitle">管理和编辑皮影戏动作排练剧本</p>
      </div>
      <div class="page-actions">
        <button class="btn-primary" @click="showCreateModal = true">
          <span class="btn-icon">+</span>
          <span>新建剧本</span>
        </button>
      </div>
    </div>

    <div class="scripts-grid">
      <div 
        v-for="script in scripts" 
        :key="script.id"
        class="script-card"
        @click="openScript(script.id)"
      >
        <div class="script-card-header">
          <div class="script-icon">🎭</div>
          <div class="script-title-section">
            <h3 class="script-name">{{ script.name }}</h3>
            <div class="script-meta-row">
              <span class="script-duration">{{ formatDuration(script.duration) }}</span>
              <span v-if="script.hasAudio" class="audio-badge" title="已添加鼓点音频">♪</span>
              <span v-if="script.beatCount > 0" class="beat-badge" :title="script.beatCount + ' 个节拍点'">🥁 {{ script.beatCount }}</span>
            </div>
          </div>
        </div>
        <p class="script-description">{{ script.description || '暂无描述' }}</p>
        <div class="script-card-footer">
          <span class="script-meta">角色: {{ getCharacterName(script.characterId) }}</span>
          <div class="script-actions" @click.stop>
            <button class="icon-btn" @click="duplicateScript(script.id)" title="复制">
              📋
            </button>
            <button class="icon-btn delete" @click="deleteScriptItem(script.id)" title="删除">
              🗑️
            </button>
          </div>
        </div>
      </div>

      <div v-if="scripts.length === 0" class="empty-state">
        <div class="empty-icon">📜</div>
        <p class="empty-text">暂无剧本，点击上方按钮创建第一个剧本</p>
      </div>
    </div>

    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h2>新建动作剧本</h2>
          <button class="modal-close" @click="showCreateModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>剧本名称</label>
            <input 
              type="text" 
              v-model="newScript.name" 
              placeholder="请输入剧本名称"
              class="form-input"
            />
          </div>
          <div class="form-group">
            <label>选择角色</label>
            <select v-model="newScript.characterId" class="form-select">
              <option value="">请选择皮影角色</option>
              <option v-for="char in characters" :key="char.id" :value="char.id">
                {{ char.name }} ({{ char.category }})
              </option>
            </select>
          </div>
          <div class="form-group">
            <label>剧本描述</label>
            <textarea 
              v-model="newScript.description" 
              placeholder="简要描述这个动作剧本"
              class="form-textarea"
              rows="3"
            ></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showCreateModal = false">取消</button>
          <button class="btn-primary" @click="createNewScript" :disabled="!canCreate">
            创建
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { getCharacters, getScripts, deleteScript, duplicateScript as dupScript } from '@/api'
import { useEditorStore } from '@/stores/editor'

const router = useRouter()
const store = useEditorStore()

const scripts = ref([])
const characters = ref([])
const showCreateModal = ref(false)
const newScript = ref({
  name: '',
  characterId: '',
  description: ''
})

const canCreate = computed(() => {
  return newScript.value.name.trim() && newScript.value.characterId
})

onMounted(async () => {
  await loadData()
})

async function loadData() {
  const [scriptsData, charsData] = await Promise.all([
    getScripts(),
    getCharacters()
  ])
  scripts.value = scriptsData
  characters.value = charsData
}

function formatDuration(seconds) {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

function getCharacterName(characterId) {
  const char = characters.value.find(c => c.id === characterId)
  return char ? char.name : '未知'
}

function openScript(id) {
  router.push(`/editor/${id}`)
}

async function createNewScript() {
  if (!canCreate.value) return
  
  const result = await store.createNewScript(
    newScript.value.characterId,
    newScript.value.name
  )
  
  showCreateModal.value = false
  newScript.value = { name: '', characterId: '', description: '' }
  
  router.push(`/editor/${result.id}`)
}

async function duplicateScript(id) {
  const result = await dupScript(id, '副本')
  await loadData()
  return result
}

async function deleteScriptItem(id) {
  if (!confirm('确定要删除这个剧本吗？')) return
  
  await store.deleteScript(id)
  await loadData()
}
</script>

<style scoped>
.script-list-page {
  height: 100%;
  overflow-y: auto;
  padding: 24px 32px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.page-title-section {
  flex: 1;
}

.page-title {
  font-size: 24px;
  font-weight: 700;
  color: #e0e0e0;
  margin-bottom: 4px;
}

.page-subtitle {
  font-size: 14px;
  color: #8080a0;
}

.page-actions {
  flex-shrink: 0;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  background: linear-gradient(135deg, #e8b059, #d4944a);
  color: #1a1a2e;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #f4c069, #e8a45a);
  transform: translateY(-1px);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-icon {
  font-size: 16px;
  line-height: 1;
}

.scripts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.script-card {
  background: #252545;
  border-radius: 10px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.25s;
  border: 1px solid #3a3a5a;
}

.script-card:hover {
  background: #2d2d50;
  border-color: #e8b059;
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
}

.script-card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.script-icon {
  font-size: 32px;
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(232, 176, 89, 0.15);
  border-radius: 8px;
}

.script-title-section {
  flex: 1;
  min-width: 0;
}

.script-name {
  font-size: 16px;
  font-weight: 600;
  color: #e0e0e0;
  margin-bottom: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.script-duration {
  font-size: 12px;
  color: #e8b059;
  font-family: 'Courier New', monospace;
}

.script-meta-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.audio-badge {
  font-size: 13px;
  color: #4ade80;
  background: rgba(74, 222, 128, 0.12);
  border-radius: 4px;
  padding: 0 6px;
  line-height: 16px;
}

.beat-badge {
  font-size: 10px;
  font-weight: 500;
  color: #4ade80;
  background: rgba(74, 222, 128, 0.15);
  border: 1px solid rgba(74, 222, 128, 0.3);
  border-radius: 10px;
  padding: 1px 7px;
}

.script-description {
  font-size: 13px;
  color: #9090b0;
  line-height: 1.5;
  margin-bottom: 16px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 39px;
}

.script-card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 12px;
  border-top: 1px solid #3a3a5a;
}

.script-meta {
  font-size: 12px;
  color: #707090;
}

.script-actions {
  display: flex;
  gap: 4px;
}

.icon-btn {
  width: 28px;
  height: 28px;
  border-radius: 4px;
  background: transparent;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}

.icon-btn:hover {
  background: #3a3a5a;
}

.icon-btn.delete:hover {
  background: rgba(220, 80, 80, 0.2);
}

.empty-state {
  grid-column: 1 / -1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #707090;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.empty-text {
  font-size: 14px;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: #252545;
  border-radius: 12px;
  width: 90%;
  max-width: 480px;
  border: 1px solid #3a3a5a;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #3a3a5a;
}

.modal-header h2 {
  font-size: 18px;
  font-weight: 600;
  color: #e0e0e0;
}

.modal-close {
  width: 32px;
  height: 32px;
  background: transparent;
  color: #8080a0;
  font-size: 24px;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.2s;
}

.modal-close:hover {
  background: #3a3a5a;
  color: #e0e0e0;
}

.modal-body {
  padding: 24px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-group label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: #b0b0d0;
  margin-bottom: 6px;
}

.form-input,
.form-select,
.form-textarea {
  width: 100%;
  padding: 10px 12px;
  background: #1a1a2e;
  border: 1px solid #3a3a5a;
  border-radius: 6px;
  color: #e0e0e0;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
}

.form-input:focus,
.form-select:focus,
.form-textarea:focus {
  border-color: #e8b059;
}

.form-select {
  cursor: pointer;
}

.form-textarea {
  resize: vertical;
  font-family: inherit;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 24px;
  border-top: 1px solid #3a3a5a;
}

.btn-secondary {
  padding: 8px 20px;
  background: #3a3a5a;
  color: #c0c0e0;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s;
}

.btn-secondary:hover {
  background: #4a4a7a;
}
</style>
