import { useEffect, useId, useRef, useState } from 'react'
import { FiChevronDown, FiSearch } from 'react-icons/fi'
import { formatNumber } from '../../../utils/format'

// options: [{ id, name, quantity, minQty }] — quantity is the on-hand count at the chosen branch (null = unknown)
function QtyBadge({ quantity, minQty }) {
  if (quantity == null) return null
  const tone =
    quantity === 0 ? 'bg-danger-bg text-danger' : quantity <= minQty ? 'bg-warn-bg text-warn' : 'bg-surface-2 text-muted'
  return (
    <span className={`shrink-0 rounded-md px-2 py-0.5 text-[11px] font-semibold tabular-nums ${tone}`}>
      คงเหลือ {formatNumber(quantity)} ชิ้น
    </span>
  )
}

// Searchable product picker that expands in place (no overlay, so the modal body never clips it)
function ProductCombobox({ id, value, options, status, invalid, disabled, onChange }) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [activeIndex, setActiveIndex] = useState(0)
  const listRef = useRef(null)
  const buttonRef = useRef(null)
  const listId = useId()

  const selected = options.find((option) => option.id === value)
  const keyword = query.trim().toLowerCase()
  const filtered = keyword ? options.filter((option) => option.name.toLowerCase().includes(keyword)) : options
  const active = Math.min(activeIndex, filtered.length - 1)
  const optionId = (index) => `${listId}-${index}`

  useEffect(() => {
    if (open) listRef.current?.querySelector(`[data-index="${active}"]`)?.scrollIntoView({ block: 'nearest' })
  }, [active, open])

  const close = () => {
    setOpen(false)
    setQuery('')
    buttonRef.current?.focus()
  }

  const pick = (option) => {
    onChange(option.id)
    close()
  }

  const onKeyDown = (event) => {
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      setActiveIndex(Math.min(active + 1, filtered.length - 1))
    } else if (event.key === 'ArrowUp') {
      event.preventDefault()
      setActiveIndex(Math.max(active - 1, 0))
    } else if (event.key === 'Enter') {
      event.preventDefault() // don't submit the form
      if (filtered[active]) pick(filtered[active])
    } else if (event.key === 'Escape') {
      event.preventDefault()
      event.stopPropagation() // close the picker only, not the modal
      close()
    }
  }

  const message =
    status === 'loading'
      ? 'กำลังโหลดรายการสินค้า…'
      : status === 'error'
        ? 'โหลดรายการสินค้าไม่สำเร็จ — ปิดหน้าต่างแล้วเปิดใหม่อีกครั้ง'
        : filtered.length === 0
          ? 'ไม่พบสินค้า'
          : null

  if (!open) {
    return (
      <button
        ref={buttonRef}
        id={id}
        type="button"
        disabled={disabled}
        aria-haspopup="listbox"
        aria-invalid={invalid ? 'true' : undefined}
        onClick={() => {
          setOpen(true)
          setActiveIndex(Math.max(0, options.findIndex((option) => option.id === value)))
        }}
        className={`flex h-11 w-full items-center gap-3 rounded-control border bg-white px-3.5 text-left text-sm transition focus-visible:ring-3 focus-visible:outline-none disabled:cursor-not-allowed disabled:bg-surface-2 ${
          invalid ? 'border-danger focus-visible:ring-danger/10' : 'border-line-strong focus-visible:ring-chart/10'
        }`}
      >
        <span className={`min-w-0 flex-1 truncate ${selected ? 'text-ink' : 'text-subtle'}`}>
          {selected?.name ?? 'เลือกสินค้า'}
        </span>
        {selected && <QtyBadge quantity={selected.quantity} minQty={selected.minQty} />}
        <FiChevronDown className="size-4 shrink-0 text-muted" />
      </button>
    )
  }

  return (
    <div className="rounded-control border border-primary/60 bg-white p-2.5 shadow-sm">
      <label className="relative block">
        <FiSearch className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-subtle" />
        <input
          autoFocus
          id={id}
          type="text"
          role="combobox"
          aria-expanded="true"
          aria-controls={listId}
          aria-autocomplete="list"
          aria-activedescendant={filtered[active] ? optionId(active) : undefined}
          aria-label="ค้นหาสินค้า"
          placeholder="ค้นหาสินค้า"
          value={query}
          onChange={(event) => {
            setQuery(event.target.value)
            setActiveIndex(0)
          }}
          onKeyDown={onKeyDown}
          onBlur={() => {
            // Option clicks preventDefault on mousedown, so a blur means focus really left the picker
            setOpen(false)
            setQuery('')
          }}
          className="h-10 w-full rounded-control border border-line-strong pr-3 pl-9 text-sm placeholder:text-subtle focus:border-chart/40 focus:ring-3 focus:ring-chart/10 focus:outline-none"
        />
      </label>

      {message ? (
        <p className="px-2 py-4 text-center text-sm text-muted">{message}</p>
      ) : (
        <ul ref={listRef} id={listId} role="listbox" aria-label="สินค้า" className="mt-2 max-h-56 overflow-y-auto">
          {filtered.map((option, index) => (
            <li
              key={option.id}
              id={optionId(index)}
              role="option"
              aria-selected={option.id === value}
              data-index={index}
              onMouseDown={(event) => event.preventDefault()} // keep focus in the search box
              onMouseEnter={() => setActiveIndex(index)}
              onClick={() => pick(option)}
              className={`flex cursor-pointer items-center justify-between gap-3 rounded-lg px-3 py-2.5 text-sm ${
                index === active ? 'bg-surface-2' : ''
              } ${option.id === value ? 'font-semibold' : ''}`}
            >
              <span className="min-w-0 text-ink [overflow-wrap:anywhere]">{option.name}</span>
              <QtyBadge quantity={option.quantity} minQty={option.minQty} />
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

export default ProductCombobox
