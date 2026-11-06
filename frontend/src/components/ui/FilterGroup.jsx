import { formatNumber } from '../../utils/format'

// A row of mutually exclusive filter buttons, each with an optional count.
// options: [{ value, label }] · counts: { [value]: number } (missing = no badge)
function FilterGroup({ label, options, value, counts, onChange }) {
  return (
    <div role="group" aria-label={label} className="flex rounded-[12px] border border-line bg-surface-2 p-1">
      {options.map((option) => (
        <button
          key={option.value || 'all'}
          type="button"
          aria-pressed={value === option.value}
          onClick={() => onChange(option.value)}
          className={`flex h-9 items-center gap-2 rounded-lg px-3.5 text-sm whitespace-nowrap transition focus-visible:ring-3 focus-visible:ring-chart/20 focus-visible:outline-none ${
            value === option.value ? 'bg-white font-semibold text-ink shadow-sm' : 'text-muted hover:text-ink'
          }`}
        >
          {option.label}
          {counts[option.value] != null && (
            <span
              className={`min-w-6 rounded-full px-1.5 py-0.5 text-center text-xs font-medium ${
                value === option.value ? 'bg-surface-2 text-ink' : 'bg-white/70 text-muted'
              }`}
            >
              {formatNumber(counts[option.value])}
            </span>
          )}
        </button>
      ))}
    </div>
  )
}

export default FilterGroup
