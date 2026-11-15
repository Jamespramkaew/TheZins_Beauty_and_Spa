import { useEffect, useId, useRef, useState } from 'react'
import { FiSearch } from 'react-icons/fi'
import { isContainer } from '../courseUtils'

const STATUS_MESSAGES = {
  loading: 'กำลังโหลดรายการสินค้า…',
  error: 'โหลดรายการสินค้าไม่สำเร็จ — ปิดฟอร์มแล้วเปิดใหม่อีกครั้ง',
}

// Searchable combobox. `stocks` is already filtered to what can be added.
function StockPicker({ stocks, status, onSelect, onClose }) {
  const [query, setQuery] = useState('')
  const [activeIndex, setActiveIndex] = useState(0)
  const listRef = useRef(null)
  const listId = useId()

  const keyword = query.trim().toLowerCase()
  const options = keyword ? stocks.filter((stock) => stock.name.toLowerCase().includes(keyword)) : stocks
  const active = Math.min(activeIndex, options.length - 1)
  const optionId = (index) => `${listId}-${index}`

  useEffect(() => {
    listRef.current?.querySelector(`[data-index="${active}"]`)?.scrollIntoView({ block: 'nearest' })
  }, [active])

  const onKeyDown = (event) => {
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      setActiveIndex(Math.min(active + 1, options.length - 1))
    } else if (event.key === 'ArrowUp') {
      event.preventDefault()
      setActiveIndex(Math.max(active - 1, 0))
    } else if (event.key === 'Enter') {
      event.preventDefault() // don't submit the course form
      if (options[active]) onSelect(options[active])
    } else if (event.key === 'Escape') {
      event.preventDefault()
      event.stopPropagation() // close the picker only, not the modal
      onClose()
    }
  }

  const message = STATUS_MESSAGES[status] ?? (options.length ? null : 'ไม่พบสินค้าที่เพิ่มได้')

  return (
    <div className="rounded-control border border-primary/60 bg-white p-2.5 shadow-sm">
      <div className="flex gap-2">
        <label className="relative flex-1">
          <FiSearch className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-subtle" />
          <input
            autoFocus
            type="text"
            role="combobox"
            aria-expanded="true"
            aria-controls={listId}
            aria-autocomplete="list"
            aria-activedescendant={options[active] ? optionId(active) : undefined}
            aria-label="ค้นหาสินค้าในสต็อก"
            placeholder="ค้นหาสินค้าในสต็อก"
            value={query}
            onChange={(event) => {
              setQuery(event.target.value)
              setActiveIndex(0)
            }}
            onKeyDown={onKeyDown}
            className="h-10 w-full rounded-control border border-line-strong pr-3 pl-9 text-sm placeholder:text-subtle focus:border-chart/40 focus:ring-3 focus:ring-chart/10 focus:outline-none"
          />
        </label>
        <button
          type="button"
          onClick={onClose}
          className="h-10 rounded-control border border-line-strong px-3 text-sm text-ink transition hover:bg-surface-2"
        >
          ปิด
        </button>
      </div>

      {message ? (
        <p className="px-2 py-4 text-center text-sm text-muted">{message}</p>
      ) : (
        <ul ref={listRef} id={listId} role="listbox" aria-label="สินค้าที่เพิ่มได้" className="mt-2 max-h-56 overflow-y-auto">
          {options.map((stock, index) => (
            <li
              key={stock.id}
              id={optionId(index)}
              role="option"
              aria-selected={index === active}
              data-index={index}
              onMouseDown={(event) => event.preventDefault()} // keep focus in the search box
              onMouseEnter={() => setActiveIndex(index)}
              onClick={() => onSelect(stock)}
              className={`flex cursor-pointer items-center justify-between gap-3 rounded-lg px-3 py-2.5 text-sm ${
                index === active ? 'bg-surface-2' : ''
              }`}
            >
              <span className="min-w-0 text-ink [overflow-wrap:anywhere]">{stock.name}</span>
              <span
                className={`shrink-0 rounded-md px-2 py-0.5 text-[11px] font-medium ${
                  isContainer(stock) ? 'bg-[#F3EDF7] text-[#5B3A7A]' : 'bg-[#EEF1F8] text-chart'
                }`}
              >
                {isContainer(stock) ? `กระปุก · ${stock.uses_per_unit} ครั้ง` : 'ชิ้นเดียว'}
              </span>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

export default StockPicker
