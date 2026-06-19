import { defineStore } from 'pinia'
import { getCharacter, getScript, updateScript, getScripts, createScript, deleteScript, duplicateScript, uploadAudio, getAudioUrl, deleteAudio, updateBeats, analyzeBeats } from '@/api'
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
    scriptsList: [],

    audioElement: null,
    audioUrl: '',
    audioLoaded: false,
    audioDuration: 0,
    isAudioPlaying: false,
    syncedKeyframes: new Map(),

    beatFlash: null
  }),

  getters: {
    duration: (state) => state.script?.duration || 10,
    fps: (state) => state.script?.fps || 30,
    tracks: (state) => state.script?.tracks || [],
    bones: (state) => state.character?.bones || [],
    audioTrack: (state) => state.script?.audioTrack || null,
    beats: (state) => state.script?.audioTrack?.beats || [],
    syncTolerance: (state) => state.script?.audioTrack?.syncTolerance || 0.08
  },

  actions: {
    async loadScriptsList() {
      this.scriptsList = await getScripts()
      return this.scriptsList
    },

    async loadScript(scriptId) {
      this.destroyAudio()
      this.script = await getScript(scriptId)
      if (this.script?.characterId) {
        this.character = await getCharacter(this.script.characterId)
      }
      this.currentTime = 0
      this.updateBonePositions()
      this.initAudioFromScript()
      this.checkAllKeyframesSynced()
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
        this.destroyAudio()
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
      this.checkBeatsAtCurrentTime()
      if (this.audioElement && this.isAudioPlaying) {
        const audioTime = time + (this.script?.audioTrack?.startTime || 0)
        if (Math.abs(this.audioElement.currentTime - audioTime) > 0.1) {
          this.audioElement.currentTime = Math.max(0, audioTime)
        }
      }
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

    initAudioFromScript() {
      if (typeof window === 'undefined') return
      const at = this.script?.audioTrack
      if (!at || !at.fileName) {
        this.destroyAudio()
        return
      }

      const url = getAudioUrl(this.script.id, at.fileName)
      this.audioUrl = url
      this.audioLoaded = false
      this.audioDuration = at.duration || this.duration

      if (this.audioElement) {
        this.audioElement.pause()
        this.audioElement.src = url
      } else {
        this.audioElement = new Audio()
        this.audioElement.src = url
        this.audioElement.crossOrigin = 'anonymous'
        this.audioElement.preload = 'auto'
      }

      if (at.volume !== undefined) {
        this.audioElement.volume = at.volume
      }

      this.audioElement.onloadedmetadata = () => {
        this.audioLoaded = true
        this.audioDuration = this.audioElement.duration
        if (this.script?.audioTrack) {
          this.script.audioTrack.duration = this.audioDuration
        }
      }

      this.audioElement.onerror = (e) => {
        console.error('Audio load error', e)
        this.audioLoaded = false
      }
    },

    destroyAudio() {
      if (this.audioElement) {
        this.audioElement.pause()
        this.audioElement.src = ''
        this.audioElement = null
      }
      this.audioUrl = ''
      this.audioLoaded = false
      this.isAudioPlaying = false
      this.audioDuration = 0
      this.syncedKeyframes.clear()
    },

    play() {
      if (this.isPlaying) return
      this.isPlaying = true
      const interval = 1000 / this.fps
      
      if (this.audioElement && this.audioLoaded) {
        const startTime = (this.script?.audioTrack?.startTime || 0) + this.currentTime
        if (Math.abs(this.audioElement.currentTime - startTime) > 0.2) {
          this.audioElement.currentTime = Math.max(0, startTime)
        }
        this.audioElement.loop = true
        this.audioElement.play().then(() => {
          this.isAudioPlaying = true
        }).catch(e => {
          console.warn('Audio play blocked:', e)
          this.isAudioPlaying = false
        })
      }

      this.playInterval = setInterval(() => {
        let newTime = this.currentTime + 1 / this.fps
        let looped = false
        if (newTime >= this.duration) {
          newTime = 0
          looped = true
        }
        this.currentTime = newTime
        this.updateBonePositions()
        if (looped && this.audioElement && this.audioLoaded && !this.isAudioPlaying) {
          this.audioElement.currentTime = this.script?.audioTrack?.startTime || 0
          this.audioElement.play().then(() => this.isAudioPlaying = true).catch(() => {})
        }
        this.checkBeatsAtCurrentTime()
      }, interval)
    },

    pause() {
      this.isPlaying = false
      if (this.playInterval) {
        clearInterval(this.playInterval)
        this.playInterval = null
      }
      if (this.audioElement) {
        this.audioElement.pause()
        this.isAudioPlaying = false
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
      this.checkAllKeyframesSynced()
    },

    deleteKeyframe(trackIndex, keyframeIndex) {
      const track = this.script.tracks[trackIndex]
      if (!track || track.keyframes.length <= 1) return
      
      track.keyframes.splice(keyframeIndex, 1)
      this.updateBonePositions()
      this.checkAllKeyframesSynced()
    },

    moveKeyframe(trackIndex, keyframeIndex, newTime) {
      const track = this.script.tracks[trackIndex]
      if (!track) return

      const clampedTime = Math.max(0, Math.min(this.duration, newTime))
      track.keyframes[keyframeIndex].time = clampedTime
      track.keyframes.sort((a, b) => a.time - b.time)
      this.updateBonePositions()
      this.checkAllKeyframesSynced()
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
    },

    async uploadAudioFile(file, onProgress) {
      if (!this.script) throw new Error('No script loaded')
      const result = await uploadAudio(this.script.id, file, onProgress)
      this.script = await getScript(this.script.id)
      this.initAudioFromScript()
      return result
    },

    async removeAudio() {
      if (!this.script) return
      await deleteAudio(this.script.id)
      this.script = await getScript(this.script.id)
      this.destroyAudio()
      this.syncedKeyframes.clear()
    },

    async saveBeats(beats, opts = {}) {
      if (!this.script) return
      const data = {
        beats,
        syncTolerance: opts.syncTolerance ?? this.script.audioTrack?.syncTolerance,
        startTime: opts.startTime ?? this.script.audioTrack?.startTime,
        volume: opts.volume ?? this.script.audioTrack?.volume
      }
      this.script = await updateBeats(this.script.id, data)
      this.checkAllKeyframesSynced()
      if (this.audioElement && this.script.audioTrack) {
        this.audioElement.volume = this.script.audioTrack.volume
      }
      return this.script
    },

    async generateBeatsFromBPM(bpm, offset = 0) {
      if (!this.script) return []
      const result = await analyzeBeats(this.script.id, bpm, offset)
      if (result.beats) {
        await this.saveBeats(result.beats)
      }
      return result
    },

    addBeat(time) {
      if (!this.script?.audioTrack) return
      const beats = [...(this.script.audioTrack.beats || [])]
      beats.push({ time, label: '', enabled: true })
      beats.sort((a, b) => a.time - b.time)
      this.script.audioTrack.beats = beats
      this.checkAllKeyframesSynced()
    },

    removeBeat(index) {
      if (!this.script?.audioTrack) return
      this.script.audioTrack.beats.splice(index, 1)
      this.checkAllKeyframesSynced()
    },

    isKeyframeSynced(trackIdx, kfIdx) {
      const key = `${trackIdx}-${kfIdx}`
      return !!this.syncedKeyframes.get(key)
    },

    getSyncedBeatFor(trackIdx, kfIdx) {
      const key = `${trackIdx}-${kfIdx}`
      return this.syncedKeyframes.get(key) || null
    },

    checkAllKeyframesSynced() {
      this.syncedKeyframes.clear()
      if (!this.script?.audioTrack?.beats?.length) return

      const beats = this.script.audioTrack.beats.filter(b => b.enabled)
      const tol = this.syncTolerance

      this.script.tracks.forEach((track, ti) => {
        track.keyframes.forEach((kf, ki) => {
          for (const beat of beats) {
            if (Math.abs(kf.time - beat.time) <= tol) {
              const key = `${ti}-${ki}`
              this.syncedKeyframes.set(key, {
                beatTime: beat.time,
                offset: kf.time - beat.time
              })
              break
            }
          }
        })
      })
    },

    checkBeatsAtCurrentTime() {
      if (!this.script?.audioTrack?.beats?.length) {
        if (this.beatFlash !== null) this.beatFlash = null
        return
      }
      const beats = this.script.audioTrack.beats.filter(b => b.enabled)
      const tol = this.syncTolerance * 0.8
      let hit = null
      for (const beat of beats) {
        if (Math.abs(this.currentTime - beat.time) < tol) {
          hit = beat.time
          break
        }
      }
      if (hit !== null && this.beatFlash !== hit) {
        this.beatFlash = hit
      } else if (hit === null) {
        this.beatFlash = null
      }
    }
  }
})
