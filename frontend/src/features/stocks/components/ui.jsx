import { FiMinus, FiPlus, FiRefreshCw, FiSearch, FiWifiOff, FiX } from 'react-icons/fi'
import { loadErrorText } from '../stockErrors'
import { outlineButtonClass } from './styles'

export function Field({ id, label, required, optional, hint, error, children }) {
  return (
    <div>
      <label htmlFor={id} className="mb-1.5 block text-sm font-medium text-ink">
        {label}
        {required && <span className="ml-0.5 text-danger">*</span>}
        {optional && <span className="ml-1 font-normal text-subtle">(ไม่บังคับ)</span>}
      </label>
      {children}
      {error ? (
        <p className="mt-1.5 text-xs text-danger">{error}</p>
      ) : (
        hint && <p className="mt-1.5 text-xs text-muted">{hint}</p>
      )}
    </div>
  )
}

// Pill-style single choice (category, direction, date preset…)
export function SegmentedControl({ label, options, value, onChange, size = 'md' }) {
  return (
    <div role="radiogroup" aria-label={label} className="flex shrink-0 rounded-[12px] border border-line bg-surface-2 p-1">
      {options.map((option) => {
        const selected = option.value === value
        return (
          <button
            key={option.value}
            type="button"
            role="radio"
            aria-checked={selected}
            onClick={() => onChange(option.value)}
            className={`rounded-lg px-3 text-sm whitespace-nowrap transition focus-visible:ring-3 focus-visible:ring-chart/20 focus-visible:outline-none ${
              size === 'sm' ? 'h-8' : 'h-9'
            } ${selected ? 'bg-white font-semibold text-ink shadow-sm' : 'text-muted hover:text-ink'}`}
          >
            {option.label}
          </button>
        )
      })}
    </div>
  )
}

export function BranchDot({ color }) {
  return <span aria-hidden="true" className="inline-block size-2 shrink-0 rounded-full" style={{ backgroundColor: color }} />
}

export function CategoryText({ category }) {
  const premium = category === 'premium'
  return (
    <span
      className={`inline-flex items-center gap-1.5 text-[11px] font-semibold tracking-wide uppercase ${
        premium ? 'text-gold-dark' : 'text-muted'
      }`}
    >
      {premium && <span aria-hidden="true" className="size-1.5 rotate-45 bg-gold" />}
      {category}
    </span>
  )
}

export function Stepper({ id, value, onChange, invalid, min = 1, label }) {
  const number = Number(value) || 0
  return (
    <div
      className={`flex h-11 w-40 items-center rounded-control border bg-white ${
        invalid ? 'border-danger ring-3 ring-danger/10' : 'border-line-strong'
      }`}
    >
      <button
        type="button"
        aria-label={`ลด${label}`}
        disabled={number <= min}
        onClick={() => onChange(Math.max(min, number - 1))}
        className="flex h-full w-11 items-center justify-center text-muted transition hover:text-ink disabled:opacity-40"
      >
        <FiMinus className="size-4" />
      </button>
      <input
        id={id}
        type="number"
        inputMode="numeric"
        min={min}
        value={value}
        onChange={(event) => onChange(event.target.value === '' ? '' : Math.max(0, Math.trunc(Number(event.target.value))))}
        aria-invalid={invalid ? 'true' : undefined}
        className={`h-full min-w-0 flex-1 bg-transparent text-center text-[15px] font-semibold tabular-nums focus:outline-none [&::-webkit-inner-spin-button]:appearance-none ${
          invalid ? 'text-danger' : 'text-ink'
        }`}
      />
      <button
        type="button"
        aria-label={`เพิ่ม${label}`}
        onClick={() => onChange(number + 1)}
        className="flex h-full w-11 items-center justify-center text-muted transition hover:text-ink"
      >
        <FiPlus className="size-4" />
      </button>
    </div>
  )
}

export function InfoBox({ icon: Icon, tone = 'neutral', children }) {
  const toneClass = tone === 'gold' ? 'bg-gold-bg text-gold-ink' : 'bg-surface-2 text-muted'
  return (
    <p className={`flex items-start gap-2.5 rounded-card px-4 py-3 text-sm ${toneClass}`}>
      <Icon className="mt-0.5 size-4 shrink-0" />
      <span>{children}</span>
    </p>
  )
}

export function NoResults({ onClear }) {
  return (
    <div className="flex flex-col items-center justify-center rounded-card border border-dashed border-line-strong bg-surface px-6 py-14 text-center">
      <span className="flex size-11 items-center justify-center rounded-full bg-white text-subtle shadow-sm">
        <FiSearch className="size-5" />
      </span>
      <p className="mt-3 font-semibold text-ink">ไม่พบตามตัวกรอง</p>
      <p className="mt-1 text-sm text-muted">ลองเปลี่ยนคำค้นหา หรือล้างตัวกรองเพื่อดูทั้งหมด</p>
      {onClear && (
        <button type="button" onClick={onClear} className={`${outlineButtonClass} mt-4 h-10`}>
          ล้างตัวกรอง
        </button>
      )}
    </div>
  )
}

export function EmptyState({ title, detail }) {
  return (
    <div className="flex flex-col items-center justify-center rounded-card border border-dashed border-line-strong bg-surface px-6 py-14 text-center">
      <p className="font-semibold text-ink">{title}</p>
      {detail && <p className="mt-1 text-sm text-muted">{detail}</p>}
    </div>
  )
}

export function LoadError({ error, onRetry }) {
  const { title, detail } = loadErrorText(error)
  return (
    <div
      role="alert"
      className="flex flex-col items-center justify-center rounded-card border border-danger/20 bg-danger-bg/50 px-6 py-14 text-center"
    >
      <span className="flex size-11 items-center justify-center rounded-full bg-white text-danger shadow-sm">
        <FiWifiOff className="size-5" />
      </span>
      <p className="mt-3 font-semibold text-ink">{title}</p>
      <p className="mt-1 text-sm text-muted">{detail}</p>
      <button type="button" onClick={onRetry} className={`${outlineButtonClass} mt-4 h-10`}>
        <FiRefreshCw className="size-4" />
        ลองใหม่
      </button>
    </div>
  )
}

export function SkeletonRows({ rows = 6 }) {
  return (
    <div aria-busy="true" aria-label="กำลังโหลด" className="overflow-hidden rounded-card border border-line">
      <div className="h-11 border-b border-line bg-surface" />
      {Array.from({ length: rows }, (_, index) => (
        <div key={index} className="flex items-center gap-6 border-b border-line px-5 py-4 last:border-0">
          <div className="h-4 w-1/4 animate-pulse rounded bg-surface-2" />
          <div className="h-4 w-1/6 animate-pulse rounded bg-surface-2" />
          <div className="ml-auto h-4 w-12 animate-pulse rounded bg-surface-2" />
          <div className="h-4 w-16 animate-pulse rounded bg-surface-2" />
        </div>
      ))}
    </div>
  )
}

export function ModalHeader({ titleId, title, subtitle, onClose }) {
  return (
    <header className="flex items-start justify-between gap-4 border-b border-line px-5 py-4 sm:px-6">
      <div>
        <h2 id={titleId} className="text-lg font-bold text-ink">
          {title}
        </h2>
        {subtitle && <p className="mt-0.5 text-[13px] text-muted">{subtitle}</p>}
      </div>
      <button
        type="button"
        onClick={onClose}
        aria-label="ปิด"
        className="flex size-10 shrink-0 items-center justify-center rounded-control bg-surface-2 text-muted transition hover:text-ink focus-visible:ring-3 focus-visible:ring-chart/20 focus-visible:outline-none"
      >
        <FiX className="size-4" />
      </button>
    </header>
  )
}
