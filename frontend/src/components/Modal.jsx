import { useEffect, useRef } from 'react'

// Esc calls onClose. Children that handle Esc themselves (e.g. an open combobox) should stopPropagation.
// placement="right" renders a full-height drawer from the right edge; clicking the backdrop closes it.
function Modal({
  onClose,
  labelledBy,
  describedBy,
  role = 'dialog',
  fullScreenOnMobile = false,
  placement = 'center',
  className = '',
  children,
}) {
  const panelRef = useRef(null)

  useEffect(() => {
    const previousFocus = document.activeElement
    const previousOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    panelRef.current?.focus()
    return () => {
      document.body.style.overflow = previousOverflow
      previousFocus?.focus?.()
    }
  }, [])

  const onKeyDown = (event) => {
    if (event.key === 'Escape') {
      event.stopPropagation()
      onClose()
    }
  }

  if (placement === 'right') {
    return (
      <div onKeyDown={onKeyDown} className="fixed inset-0 z-40 flex justify-end bg-ink/45">
        <div className="flex-1" onClick={onClose} aria-hidden="true" />
        <div
          ref={panelRef}
          role={role}
          aria-modal="true"
          aria-labelledby={labelledBy}
          aria-describedby={describedBy}
          tabIndex={-1}
          className={`flex h-full w-full flex-col bg-white shadow-[0_24px_64px_-24px_rgba(23,24,28,0.45)] outline-none ${className}`}
        >
          {children}
        </div>
      </div>
    )
  }

  return (
    <div
      onKeyDown={onKeyDown}
      className={`fixed inset-0 z-40 flex justify-center bg-ink/45 ${
        fullScreenOnMobile ? 'items-stretch sm:items-center sm:p-6' : 'items-center p-4'
      }`}
    >
      <div
        ref={panelRef}
        role={role}
        aria-modal="true"
        aria-labelledby={labelledBy}
        aria-describedby={describedBy}
        tabIndex={-1}
        className={`flex w-full flex-col bg-white shadow-[0_24px_64px_-24px_rgba(23,24,28,0.45)] outline-none ${
          fullScreenOnMobile
            ? 'h-full sm:h-auto sm:max-h-[calc(100dvh-3rem)] sm:rounded-[18px]'
            : 'max-h-[calc(100dvh-2rem)] rounded-[18px]'
        } ${className}`}
      >
        {children}
      </div>
    </div>
  )
}

export default Modal
