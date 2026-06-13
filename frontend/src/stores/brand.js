import { defineStore } from 'pinia'
import { ref } from 'vue'

const STORAGE_KEY = 'inotify_brand'

function load() {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}')
  } catch {
    return {}
  }
}

export const useBrandStore = defineStore('brand', () => {
  const saved = load()
  const icon = ref(saved.icon || '🔔')
  const name = ref(saved.name || 'Inotify')

  function save(newIcon, newName) {
    icon.value = newIcon || '🔔'
    name.value = newName || 'Inotify'
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ icon: icon.value, name: name.value }))
  }

  return { icon, name, save }
})
