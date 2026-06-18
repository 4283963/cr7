export function degToRad(deg) {
  return deg * Math.PI / 180
}

export function radToDeg(rad) {
  return rad * 180 / Math.PI
}

export function lerp(a, b, t) {
  return a + (b - a) * t
}

export function easeInOut(t) {
  return t < 0.5 ? 2 * t * t : -1 + (4 - 2 * t) * t
}

export function interpolateKeyframes(keyframes, time, interpType = 'linear') {
  if (!keyframes || keyframes.length === 0) return null
  if (keyframes.length === 1) return keyframes[0].values

  const sorted = [...keyframes].sort((a, b) => a.time - b.time)

  if (time <= sorted[0].time) return sorted[0].values
  if (time >= sorted[sorted.length - 1].time) return sorted[sorted.length - 1].values

  for (let i = 0; i < sorted.length - 1; i++) {
    const k1 = sorted[i]
    const k2 = sorted[i + 1]
    if (time >= k1.time && time <= k2.time) {
      const t = (time - k1.time) / (k2.time - k1.time)
      const easedT = interpType === 'ease' ? easeInOut(t) : t
      
      const result = {}
      for (const key of Object.keys(k1.values)) {
        result[key] = lerp(k1.values[key], k2.values[key], easedT)
      }
      return result
    }
  }

  return sorted[sorted.length - 1].values
}

export function buildBoneHierarchy(bones) {
  const boneMap = new Map()
  const childrenMap = new Map()

  for (const bone of bones) {
    boneMap.set(bone.id, bone)
    if (!childrenMap.has(bone.id)) {
      childrenMap.set(bone.id, [])
    }
    if (bone.parentId && bone.parentId !== '') {
      if (!childrenMap.has(bone.parentId)) {
        childrenMap.set(bone.parentId, [])
      }
      childrenMap.get(bone.parentId).push(bone.id)
    }
  }

  return { boneMap, childrenMap }
}

export function calculateBonePositions(bones, angleOverrides = {}, baseX = 0, baseY = 0) {
  const { boneMap, childrenMap } = buildBoneHierarchy(bones)
  const positions = {}

  function traverse(boneId, parentX, parentY, parentAngle) {
    const bone = boneMap.get(boneId)
    if (!bone) return

    const angle = (angleOverrides[boneId] !== undefined ? angleOverrides[boneId] : bone.baseAngle)
    const totalAngle = parentAngle + angle - bone.baseAngle

    const startX = parentX + bone.x
    const startY = parentY + bone.y

    const rad = degToRad(totalAngle)
    const endX = startX + Math.cos(rad) * bone.length
    const endY = startY + Math.sin(rad) * bone.length

    positions[boneId] = {
      ...bone,
      startX,
      startY,
      endX,
      endY,
      currentAngle: totalAngle,
      worldAngle: totalAngle
    }

    const children = childrenMap.get(boneId) || []
    for (const childId of children) {
      traverse(childId, endX, endY, totalAngle)
    }
  }

  const rootBone = bones.find(b => b.parentId === '' || !b.parentId)
  if (rootBone) {
    traverse(rootBone.id, baseX + rootBone.x, baseY + rootBone.y, 0)
  }

  return positions
}

export function formatTime(seconds) {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  const ms = Math.floor((seconds % 1) * 100)
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}.${ms.toString().padStart(2, '0')}`
}
