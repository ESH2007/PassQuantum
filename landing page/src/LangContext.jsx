import { createContext, useContext, useEffect, useState } from 'react'

const LangContext = createContext()
const STORAGE_KEY = 'pq-lang'
const SUPPORTED = ['es', 'en']

// Saved choice → browser language → Spanish.
function initialLang() {
  if (typeof window === 'undefined') return 'es'
  const saved = localStorage.getItem(STORAGE_KEY)
  if (SUPPORTED.includes(saved)) return saved
  return navigator.language?.startsWith('en') ? 'en' : 'es'
}

export function LangProvider({ children }) {
  const [lang, setLang] = useState(initialLang)

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, lang)
    document.documentElement.lang = lang
  }, [lang])

  return (
    <LangContext.Provider value={{ lang, setLang }}>
      {children}
    </LangContext.Provider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export function useLang() {
  return useContext(LangContext)
}
