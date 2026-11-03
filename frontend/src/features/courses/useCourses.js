import { useCallback, useEffect, useMemo, useState } from 'react'
import { courseApi } from '../../api/courseApi'

const ALL_PAGE_SIZE = 100 // backend max

async function fetchAllCourses() {
  const all = []
  for (let page = 1; ; page++) {
    const response = await courseApi.getAll({ page, page_size: ALL_PAGE_SIZE })
    all.push(...response.data)
    if (page >= (response.meta?.total_pages ?? 1)) return all
  }
}

const replaceIn = (courses, updated) => courses.map((course) => (course.id === updated.id ? updated : course))

// One page of courses for the list, plus every course for the KPI row and category suggestions.
// While a request is in flight the previous data stays on screen (loading = true).
export function useCourses({ search, page, pageSize }) {
  const [reloadKey, setReloadKey] = useState(0)
  const listKey = JSON.stringify([search, page, pageSize, reloadKey])
  const [list, setList] = useState({ key: null, courses: [], meta: null, error: null })
  const [all, setAll] = useState({ key: null, courses: null })

  useEffect(() => {
    let ignore = false
    courseApi
      .getAll({ search, page, page_size: pageSize })
      .then((response) => {
        if (!ignore) setList({ key: listKey, courses: response.data, meta: response.meta, error: null })
      })
      .catch((error) => {
        if (!ignore) setList((current) => ({ ...current, key: listKey, error }))
      })
    return () => {
      ignore = true
    }
  }, [search, page, pageSize, listKey])

  useEffect(() => {
    let ignore = false
    fetchAllCourses()
      .then((courses) => {
        if (!ignore) setAll({ key: reloadKey, courses })
      })
      .catch(() => {
        if (!ignore) setAll({ key: reloadKey, courses: null })
      })
    return () => {
      ignore = true
    }
  }, [reloadKey])

  const stats = useMemo(() => {
    if (!all.courses) return null
    const active = all.courses.filter((course) => course.is_active).length
    return {
      total: all.courses.length,
      active,
      inactive: all.courses.length - active,
      noExpiry: all.courses.filter((course) => course.valid_day == null).length,
      categories: [...new Set(all.courses.map((course) => course.category).filter(Boolean))],
    }
  }, [all.courses])

  const reload = useCallback(() => setReloadKey((key) => key + 1), [])

  // Apply a PATCH response in place, without refetching the page
  const replaceCourse = useCallback((updated) => {
    setList((current) => ({ ...current, courses: replaceIn(current.courses, updated) }))
    setAll((current) => (current.courses ? { ...current, courses: replaceIn(current.courses, updated) } : current))
  }, [])

  return {
    courses: list.courses,
    meta: list.meta,
    error: list.error,
    loading: list.key !== listKey,
    loaded: list.key !== null,
    stats,
    reload,
    replaceCourse,
  }
}
