import { defineStore } from 'pinia'
import { getCharacter, getScript, updateScript, getScripts, createScript, deleteScript, duplicateScript } from '@/api'
import { interpolateKeyframes, calculateBonePositions } from '@/utils/animation'

export const useEditorStore = defineStore('editor', {
  state: () => ({
    character: null,
    script: null,
    currentTime: 0,
    isPlaying: false,
    selectedBoneId: null,
    selectedKeyframe: null,
    playInterval: null,
    bonePositions: {},
    scriptsList: []
  }),

  getters: {
    duration: (state) => state.script?.duration || 10,
    fps: (state) => state.script?.fps || 30,
    tracks: (state) => state.script?.tracks || [],
    bones: (state) => state.character?.bones || []
  },

  actions: {
    async loadScriptsList() {
      this.scriptsList = await getScripts()
      return this.scriptsList
    },

    async loadScript(scriptId) {
      this.script = await getScript(scriptId)
      if (this.script?.characterId) {
        this.character = await getCharacter(this.script.characterId)
      }
      this.currentTime = 0
      this.updateBonePositions()
      return this.script
    },

    async createNewScript(characterId, name) {
      const character = await getCharacter(characterId)
      const tracks = character.bones
        .filter(b => b.id !== 'root')
        .map(b => ({
          boneId: b.id,
          boneName: b.name,
          keyframes: [{ time: 0, values: { angle: b.baseAngle } }],
          interpType: 'linear'
        }))
      
      const newScript = {
        name,
        description: '',
        characterId,
        duration: 10,
        fps: 30,
        tracks
      }
      
      const result = await createScript(newScript)
      this.script = result
      this.character = character
      this.currentTime = 0
      this.updateBonePositions()
      return result
    },

    async saveScript() {
      if (!this.script) return
      const result = await updateScript(this.script.id, this.script)
      this.script = result
      return result
    },

    async deleteScript(id) {
      await deleteScript(id)
      if (this.script?.id === id) {
        this.script = null
        this.character = null
      }
    },

    async duplicateScript(id, name) {
      return await duplicateScript(id, name)
    },

    setTime(time) {
      this.currentTime = Math.max(0, Math.min(this.duration, time))
      this.updateBonePositions()
    },

    updateBonePositions() {
      if (!this.character || !this.script) return

      const angleOverrides = {}
      for (const track of this.script.tracks) {
        const values = interpolateKeyframes(track.keyframes, this.currentTime, track.interpType)
        if (values && values.angle !== undefined) {
          angleOverrides[track.boneId] = values.angle
        }
      }

      this.bonePositions = calculateBonePositions(
        this.character.bones,
        angleOverrides,
        this.character.baseX,
        this.character.baseY
      )
    },

    play() {
      if (this.isPlaying) return
      this.isPlaying = true
      const interval = 1000 / this.fps
      
      this.playInterval = setInterval(() => {
        let newTime = this.currentTime + 1 / this.fps
        if (newTime >= this.duration) {
          newTime = 0
        }
        this.setTime(newTime)
      }, interval)
    },

    pause() {
      this.isPlaying = false
      if (this.playInterval) {
        clearInterval(this.playInterval)
        this.playInterval = null
      }
    },

    togglePlay() {
      if (this.isPlaying) {
        this.pause()
      } else {
        this.play()
      }
    },

    stop() {
      this.pause()
      this.setTime(0)
    },

    selectBone(boneId) {
      this.selectedBoneId = boneId
    },

    addKeyframe(trackIndex, time, angle) {
      const track = this.script.tracks[trackIndex]
      if (!track) return

      const existingIdx = track.keyframes.findIndex(k => Math.abs(k.time - time) < 0.01)
      
      if (existingIdx >= 0) {
        track.keyframes[existingIdx].values.angle = angle
      } else {
        track.keyframes.push({ time, values: { angle } })
        track.keyframes.sort((a, b) => a.time - b.time)
      }
      
      this.updateBonePositions()
    },

    deleteKeyframe(trackIndex, keyframeIndex) {
      const track = this.script.tracks[trackIndex]
      if (!track || track.keyframes.length <= 1) return
      
      track.keyframes.splice(keyframeIndex, 1)
      this.updateBonePositions()
    },

    moveKeyframe(trackIndex, keyframeIndex, newTime) {
      const track = this.script.tracks[trackIndex]
      if (!track) return

      const clampedTime = Math.max(0, Math.min(this.duration, newTime))
      track.keyframes[keyframeIndex].time = clampedTime
      track.keyframes.sort((a, b) => a.time - b.time)
      this.updateBonePositions()
    },

    setBoneAngle(boneId, angle) {
      const track = this.script.tracks.find(t => t.boneId === boneId)
      if (!track) return

      const trackIdx = this.script.tracks.indexOf(track)
      this.addKeyframe(trackIdx, this.currentTime, angle)
    },

    getBoneAngle(boneId) {
      if (this.bonePositions[boneId]) {
        return this.bonePositions[boneId].currentAngle
      }
      return 0
    },

    setDuration(duration) {
      if (this.script) {
        this.script.duration = Math.max(1, duration)
      }
    }
  }
})
