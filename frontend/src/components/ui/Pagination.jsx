import { FiChevronLeft, FiChevronRight } from 'react-icons/fi'
import { formatNumber } from '../../utils/format'

const WINDOW = 5

function Pagination({ meta, pageSizeOptions, onPageChange, onPageSizeChange, unit }) {
  const { page, page_size: pageSize, total, total_pages: totalPages } = meta
  const from = total === 0 ? 0 : (page - 1) * pageSize + 1
  const to = Math.min(page * pageSize, total)

  const start = Math.max(1, Math.min(page - Math.floor(WINDOW / 2), totalPages - WINDOW + 1))
  const pages = Array.from({ length: Math.min(WINDOW, totalPages) }, (_, index) => start + index)

  return (
    <div className="flex flex-wrap items-center justify-between gap-3 text-sm text-muted">
      <p>
        แสดง {formatNumber(from)}–{formatNumber(to)} จาก {formatNumber(total)} {unit}
      </p>

      <div className="flex items-center gap-2">
        <label className="flex items-center gap-2">
          ต่อหน้า
          <select
            value={pageSize}
            onChange={(event) => onPageSizeChange(Number(event.target.value))}
            className="h-9 rounded-control border border-line-strong bg-white px-2 text-ink"
          >
            {pageSizeOptions.map((size) => (
              <option key={size} value={size}>
                {size}
              </option>
            ))}
          </select>
        </label>

        <nav aria-label="เปลี่ยนหน้า" className="flex items-center gap-1">
          <PageButton label="หน้าก่อนหน้า" disabled={page <= 1} onClick={() => onPageChange(page - 1)}>
            <FiChevronLeft className="size-4" />
          </PageButton>
          {pages.map((number) => (
            <PageButton key={number} current={number === page} onClick={() => onPageChange(number)} label={`หน้า ${number}`}>
              {number}
            </PageButton>
          ))}
          <PageButton label="หน้าถัดไป" disabled={page >= totalPages} onClick={() => onPageChange(page + 1)}>
            <FiChevronRight className="size-4" />
          </PageButton>
        </nav>
      </div>
    </div>
  )
}

function PageButton({ label, current, disabled, onClick, children }) {
  return (
    <button
      type="button"
      aria-label={label}
      aria-current={current ? 'page' : undefined}
      disabled={disabled}
      onClick={onClick}
      className={`flex size-9 items-center justify-center rounded-control border text-sm transition disabled:opacity-40 ${
        current ? 'border-ink bg-ink font-semibold text-white' : 'border-line-strong bg-white text-ink hover:bg-surface-2'
      }`}
    >
      {children}
    </button>
  )
}

export default Pagination
