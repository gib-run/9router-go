export type Locale = 'pt-BR' | 'en'

export const language = $state<{ current: Locale }>({ current: 'en' })

export function initializeLocale() {
  if (typeof localStorage === 'undefined') return
  const saved = localStorage.getItem('9router-locale')
  if (saved === 'pt-BR' || saved === 'en') language.current = saved
  else language.current = navigator.language.toLowerCase().startsWith('pt') ? 'pt-BR' : 'en'
}

export function setLocale(value: Locale) {
  language.current = value
  if (typeof localStorage !== 'undefined') localStorage.setItem('9router-locale', value)
}
