import { useCallback, useEffect, useState } from 'react'

// Runs `fetcher(params)` whenever params (or refreshKey) change. `fetcher` must be stable (a module-level API function).
// While a request is in flight the previous data stays on screen (loading = true), like useCourses.
// Responses that arrive after params changed again are ignored.
export function useQuery(fetcher, params, { enabled = true, refreshKey = 0 } = {}) {
  const [reloadKey, setReloadKey] = useState(0)
  const key = JSON.stringify([params, refreshKey, reloadKey])
  const [state, setState] = useState({ key: null, data: null, meta: null, error: null, loadedAt: null })

  useEffect(() => {
    if (!enabled) return undefined
    let ignore = false
    const [requestParams] = JSON.parse(key)
    fetcher(requestParams)
      .then((response) => {
        if (!ignore) {
          setState({ key, data: response.data ?? response, meta: response.meta ?? null, error: null, loadedAt: new Date() })
        }
      })
      .catch((error) => {
        if (!ignore) setState((current) => ({ ...current, key, error }))
      })
    return () => {
      ignore = true
    }
  }, [fetcher, key, enabled])

  const reload = useCallback(() => setReloadKey((value) => value + 1), [])

  return {
    data: state.data,
    meta: state.meta,
    error: state.error,
    loadedAt: state.loadedAt,
    loading: enabled && state.key !== key,
    loaded: state.key !== null && state.data !== null,
    reload,
  }
}
