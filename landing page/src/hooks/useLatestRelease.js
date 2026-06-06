// src/hooks/useLatestRelease.js
import { useState, useEffect } from 'react'

const CACHE_KEY = 'pq_latest_release'
const API_URL   = 'https://api.github.com/repos/ESH2007/PassQuantum/releases/latest'

let _promise = null

export function useLatestRelease() {
  const [release, setRelease] = useState(() => {
    try {
      const raw = sessionStorage.getItem(CACHE_KEY)
      return raw ? JSON.parse(raw) : null
    } catch { return null }
  })

  useEffect(() => {
    if (release) return
    if (!_promise) {
      _promise = fetch(API_URL)
        .then(r => (r.ok ? r.json() : null))
        .catch(() => null)
    }
    _promise.then(data => {
      if (!data) return
      const entry = { tag: data.tag_name, name: data.name }
      try { sessionStorage.setItem(CACHE_KEY, JSON.stringify(entry)) } catch {} // eslint-disable-line no-empty
      setRelease(entry)
    })
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  return release // { tag: "v1.2.0", name: "..." } | null
}
