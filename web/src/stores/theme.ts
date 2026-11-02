import { defineStore } from 'pinia'
import { ref } from 'vue'

const KEY = 'rs_theme'
type Theme = 'light' | 'dark'

export const useThemeStore = defineStore('theme', () => {
  const saved = (localStorage.getItem(KEY) as Theme) || 'light'
  const theme = ref<Theme>(saved)

  function apply() {
    document.documentElement.dataset.theme = theme.value
  }

  function toggle() {
    theme.value = theme.value === 'light' ? 'dark' : 'light'
    localStorage.setItem(KEY, theme.value)
    apply()
  }

  apply()

  return { theme, toggle }
})
