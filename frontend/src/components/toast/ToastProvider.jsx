import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { LuCircleCheck, LuCircleX, LuX } from 'react-icons/lu'
import { ToastContext } from './ToastContext'

const DURATION_MS = 3500
const EXIT_MS = 200
const MAX_TOASTS = 3

const variants = {
  success: {
    icon: LuCircleCheck,
    iconClass: 'bg-success-bg text-success',
    barClass: 'bg-success',
    role: 'status',
    defaultTitle: 'บันทึกสำเร็จ',
  },
  error: {
    icon: LuCircleX,
    iconClass: 'bg-danger-bg text-danger',
    barClass: 'bg-danger',
    role: 'alert',
    defaultTitle: 'เกิดข้อผิดพลาด',
    defaultDescription: 'กรุณาลองใหม่อีกครั้ง',
  },
}

function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([])
  const nextId = useRef(0)
  const timers = useRef(new Map())

  const dismiss = useCallback((id) => {
    clearTimeout(timers.current.get(id))
    timers.current.delete(id)
    setToasts((current) => current.map((toast) => (toast.id === id ? { ...toast, leaving: true } : toast)))
    setTimeout(() => setToasts((current) => current.filter((toast) => toast.id !== id)), EXIT_MS)
  }, [])

  const show = useCallback(
    (type, title, description) => {
      const id = ++nextId.current
      const { defaultTitle, defaultDescription } = variants[type]
      const toast = { id, type, title: title ?? defaultTitle, description: description ?? (title ? undefined : defaultDescription) }
      setToasts((current) => [...current, toast].slice(-MAX_TOASTS))
      timers.current.set(id, setTimeout(() => dismiss(id), DURATION_MS))
    },
    [dismiss],
  )

  useEffect(() => {
    const pending = timers.current
    return () => pending.forEach(clearTimeout)
  }, [])

  const api = useMemo(
    () => ({
      success: (title, description) => show('success', title, description),
      error: (title, description) => show('error', title, description),
    }),
    [show],
  )

  return (
    <ToastContext.Provider value={api}>
      {children}
      <div
        aria-live="polite"
        className="pointer-events-none fixed right-4 bottom-4 left-4 z-50 flex flex-col items-end gap-2 sm:left-auto lg:right-8 lg:bottom-6"
      >
        {toasts.map((toast) => (
          <ToastItem key={toast.id} toast={toast} onClose={() => dismiss(toast.id)} />
        ))}
      </div>
    </ToastContext.Provider>
  )
}

function ToastItem({ toast, onClose }) {
  const { icon: Icon, iconClass, barClass, role } = variants[toast.type]

  return (
    <div
      role={role}
      className={`pointer-events-auto relative w-full overflow-hidden rounded-card border border-line bg-white shadow-[0_16px_40px_-16px_rgba(23,24,28,0.35)] sm:w-90 ${
        toast.leaving ? 'motion-safe:animate-toast-out motion-reduce:opacity-0' : 'motion-safe:animate-toast-in'
      }`}
    >
      <div className="flex min-h-22 items-start gap-3.5 px-5 py-5 pr-12">
        <span className={`flex size-10 shrink-0 items-center justify-center rounded-full ${iconClass}`}>
          <Icon className="size-5" />
        </span>
        <div className="min-w-0 pt-0.5">
          <p className="text-[15px] font-semibold text-ink">{toast.title}</p>
          {toast.description && (
            <p className="mt-1 text-[13px] leading-relaxed break-words text-muted">{toast.description}</p>
          )}
        </div>
      </div>
      <button
        type="button"
        onClick={onClose}
        aria-label="ปิดการแจ้งเตือน"
        className="absolute top-4 right-3 flex size-8 items-center justify-center rounded-lg text-subtle transition hover:bg-surface-2 hover:text-ink"
      >
        <LuX className="size-4" />
      </button>
      {/* Time left before auto-dismiss */}
      <span
        className={`absolute bottom-0 left-0 h-0.75 w-full origin-left ${barClass}`}
        style={{ animation: `toast-progress ${DURATION_MS}ms linear forwards` }}
      />
    </div>
  )
}

export default ToastProvider
