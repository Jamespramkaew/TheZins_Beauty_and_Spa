import { useEffect, useId, useLayoutEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { FiMoreHorizontal } from 'react-icons/fi'

const MENU_WIDTH = 232

// ⋯ button + popup menu. The popup is portalled to <body> and position:fixed: inside the table it would be
// clipped by the overflow box, and a hovered row's `filter` would turn fixed into row-relative positioning.
// items: [{ label, icon, onSelect } | 'divider']
function RowMenu({ label, items }) {
  const [open, setOpen] = useState(false)
  const [position, setPosition] = useState(null)
  const buttonRef = useRef(null)
  const menuRef = useRef(null)
  const menuId = useId()

  useLayoutEffect(() => {
    if (!open) return
    const rect = buttonRef.current.getBoundingClientRect()
    const menuHeight = menuRef.current?.offsetHeight ?? 0
    const below = rect.bottom + 6
    const top = below + menuHeight > window.innerHeight - 8 ? Math.max(8, rect.top - 6 - menuHeight) : below
    setPosition({ top, left: Math.max(8, rect.right - MENU_WIDTH) })
  }, [open])

  // Focus the first item once the menu is in place; preventScroll so focusing can't fire the scroll-to-close listener
  useEffect(() => {
    if (open && position) menuRef.current?.querySelector('[role="menuitem"]')?.focus({ preventScroll: true })
  }, [open, position])

  useEffect(() => {
    if (!open) return undefined
    const close = () => setOpen(false)
    const onPointerDown = (event) => {
      if (!menuRef.current?.contains(event.target) && !buttonRef.current?.contains(event.target)) close()
    }
    document.addEventListener('pointerdown', onPointerDown)
    window.addEventListener('resize', close)
    window.addEventListener('scroll', close, true)
    return () => {
      document.removeEventListener('pointerdown', onPointerDown)
      window.removeEventListener('resize', close)
      window.removeEventListener('scroll', close, true)
    }
  }, [open])

  const onMenuKeyDown = (event) => {
    const options = [...menuRef.current.querySelectorAll('[role="menuitem"]')]
    const index = options.indexOf(document.activeElement)
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      options[(index + 1) % options.length]?.focus()
    } else if (event.key === 'ArrowUp') {
      event.preventDefault()
      options[(index - 1 + options.length) % options.length]?.focus()
    } else if (event.key === 'Escape' || event.key === 'Tab') {
      event.preventDefault()
      setOpen(false)
      buttonRef.current?.focus()
    }
  }

  return (
    <>
      <button
        ref={buttonRef}
        type="button"
        aria-label={label}
        aria-haspopup="menu"
        aria-expanded={open}
        aria-controls={open ? menuId : undefined}
        onClick={() => {
          setPosition(null)
          setOpen((value) => !value)
        }}
        className={`flex size-10 items-center justify-center rounded-control text-muted transition hover:bg-surface-2 hover:text-ink focus-visible:ring-3 focus-visible:ring-chart/20 focus-visible:outline-none ${
          open ? 'border border-line-strong bg-white text-ink' : ''
        }`}
      >
        <FiMoreHorizontal className="size-4" />
      </button>

      {open &&
        createPortal(
          <div
            ref={menuRef}
            id={menuId}
            role="menu"
            aria-label={label}
            onKeyDown={onMenuKeyDown}
            style={{
              top: position?.top ?? -9999,
              left: position?.left ?? -9999,
              width: MENU_WIDTH,
            }}
            className="fixed z-30 rounded-card border border-line bg-white p-1.5 shadow-[0_16px_40px_-16px_rgba(23,24,28,0.35)]"
          >
            {items.map((item, index) =>
              item === 'divider' ? (
                <div key={`divider-${index}`} role="separator" className="my-1.5 border-t border-line" />
              ) : (
                <button
                  key={item.label}
                  type="button"
                  role="menuitem"
                  onClick={() => {
                    setOpen(false)
                    item.onSelect()
                  }}
                  className="flex h-10 w-full items-center gap-3 rounded-lg px-3 text-left text-sm text-ink transition hover:bg-surface-2 focus-visible:bg-surface-2 focus-visible:outline-none"
                >
                  <item.icon className="size-4 text-muted" />
                  {item.label}
                </button>
              ),
            )}
          </div>,
          document.body,
        )}
    </>
  )
}

export default RowMenu
