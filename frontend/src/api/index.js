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
