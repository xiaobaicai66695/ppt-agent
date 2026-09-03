import { computed, ref, watch } from 'vue'

export type ColorTheme = 'dark' | 'light'
const storageKey = 'deckform_theme'
const saved = localStorage.getItem(storageKey)
const theme = ref<ColorTheme>(saved === 'light' ? 'light' : 'dark')

function apply(value: ColorTheme) {
  document.documentElement.dataset.theme = value
  document.documentElement.style.colorScheme = value
}

apply(theme.value)
watch(theme, value => { localStorage.setItem(storageKey, value); apply(value) })

export function useTheme() {
  return { theme, isLight: computed(() => theme.value === 'light'), toggleTheme: () => { theme.value = theme.value === 'dark' ? 'light' : 'dark' } }
}
