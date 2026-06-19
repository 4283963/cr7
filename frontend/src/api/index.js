const BASE_URL = '/api'

async function request(url, options = {}) {
  const response = await fetch(BASE_URL + url, {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers
    },
    ...options
  })
  
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`)
  }
  
  return response.json()
}

export function getCharacters() {
  return request('/characters')
}

export function getCharacter(id) {
  return request(`/characters/${id}`)
}

export function createCharacter(data) {
  return request('/characters', {
    method: 'POST',
    body: JSON.stringify(data)
  })
}

export function updateCharacter(id, data) {
  return request(`/characters/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data)
  })
}

export function deleteCharacter(id) {
  return request(`/characters/${id}`, {
    method: 'DELETE'
  })
}

export function getScripts() {
  return request('/scripts')
}

export function getScript(id) {
  return request(`/scripts/${id}`)
}

export function createScript(data) {
  return request('/scripts', {
    method: 'POST',
    body: JSON.stringify(data)
  })
}

export function updateScript(id, data) {
  return request(`/scripts/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data)
  })
}

export function deleteScript(id) {
  return request(`/scripts/${id}`, {
    method: 'DELETE'
  })
}

export function duplicateScript(id, name) {
  return request(`/scripts/${id}/duplicate`, {
    method: 'POST',
    body: JSON.stringify({ name })
  })
}

export function uploadAudio(scriptId, file, onProgress) {
  return new Promise((resolve, reject) => {
    const formData = new FormData()
    formData.append('audio', file)

    const xhr = new XMLHttpRequest()
    xhr.open('POST', BASE_URL + `/scripts/${scriptId}/audio`)
    xhr.upload.onprogress = (e) => {
      if (onProgress && e.lengthComputable) {
        onProgress(e.loaded / e.total)
      }
    }
    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve(JSON.parse(xhr.responseText))
      } else {
        reject(new Error(`Upload failed: ${xhr.status}`))
      }
    }
    xhr.onerror = () => reject(new Error('Upload failed'))
    xhr.send(formData)
  })
}

export function getAudioUrl(scriptId, fileName) {
  return BASE_URL + `/scripts/${scriptId}/audio/${fileName}`
}

export function deleteAudio(scriptId) {
  return request(`/scripts/${scriptId}/audio`, {
    method: 'DELETE'
  })
}

export function updateBeats(scriptId, beatsData) {
  return request(`/scripts/${scriptId}/beats`, {
    method: 'PUT',
    body: JSON.stringify(beatsData)
  })
}

export function analyzeBeats(scriptId, bpm, offset = 0) {
  return request(`/scripts/${scriptId}/analyze-beats`, {
    method: 'POST',
    body: JSON.stringify({ bpm, offset })
  })
}
