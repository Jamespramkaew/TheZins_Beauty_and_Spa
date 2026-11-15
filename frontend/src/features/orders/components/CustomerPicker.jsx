import { useEffect, useState } from 'react'
import { FiSearch } from 'react-icons/fi'
import { customerApi } from '../../../api/customerApi'
import { useQuery } from '../../../hooks/useQuery'

// Search box + result list for choosing who the bill is for
function CustomerPicker({ onPick }) {
  const [input, setInput] = useState('')
  const [search, setSearch] = useState('')
  const results = useQuery(customerApi.search, { q: search || undefined, page: 1, page_size: 8 })

  useEffect(() => {
    const timer = setTimeout(() => setSearch(input.trim()), 300)
    return () => clearTimeout(timer)
  }, [input])

  const customers = results.data ?? []

  return (
    <div className="space-y-3">
      <label className="relative block">
        <FiSearch className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-subtle" />
        <input
          type="search"
          value={input}
          onChange={(event) => setInput(event.target.value)}
          placeholder="ค้นหาลูกค้าจากชื่อ HN หรือเบอร์โทร"
          aria-label="ค้นหาลูกค้า"
          autoFocus
          className="h-11 w-full rounded-control border border-line-strong bg-white pr-3 pl-10 text-sm placeholder:text-subtle focus:border-chart/40 focus:ring-3 focus:ring-chart/10 focus:outline-none"
        />
      </label>

      {results.error ? (
        <p className="text-sm text-danger">ค้นหาลูกค้าไม่สำเร็จ กรุณาลองใหม่</p>
      ) : !results.loaded ? (
        <p className="text-sm text-muted">กำลังค้นหา…</p>
      ) : customers.length === 0 ? (
        <p className="text-sm text-muted">{search ? 'ไม่พบลูกค้า — เพิ่มลูกค้าใหม่ได้ที่หน้าลูกค้า' : 'ยังไม่มีลูกค้า'}</p>
      ) : (
        <ul className="divide-y divide-line overflow-hidden rounded-control border border-line">
          {customers.map((customer) => (
            <li key={customer.id}>
              <button
                type="button"
                onClick={() => onPick(customer)}
                className="flex w-full items-center justify-between gap-3 bg-white px-4 py-3 text-left text-sm transition hover:bg-surface-2"
              >
                <span className="font-medium text-ink [overflow-wrap:anywhere]">
                  {[customer.fname, customer.lname].filter(Boolean).join(' ')}
                </span>
                <span className="shrink-0 text-xs text-muted">
                  {customer.hn_no}
                  {customer.tel && ` · ${customer.tel}`}
                </span>
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

export default CustomerPicker
