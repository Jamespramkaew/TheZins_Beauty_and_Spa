// Shared Tailwind class strings for the stock screens (kept out of .jsx so fast refresh works)
export const inputClass = (error) =>
  `h-11 w-full rounded-control border px-3.5 text-sm text-ink placeholder:text-subtle focus:ring-3 focus:outline-none disabled:bg-surface-2 disabled:text-muted ${
    error
      ? 'border-danger bg-danger-bg/30 focus:ring-danger/10'
      : 'border-line-strong bg-white focus:border-chart/40 focus:ring-chart/10'
  }`

export const filterSelectClass =
  'h-10 rounded-control border border-line-strong bg-white pr-8 pl-3 text-sm text-ink focus:border-chart/40 focus:ring-3 focus:ring-chart/10 focus:outline-none'

export const primaryButtonClass =
  'flex h-11 items-center justify-center gap-2 rounded-control bg-primary px-4 text-sm font-semibold text-white shadow-sm transition hover:bg-[#4d1621] focus-visible:ring-3 focus-visible:ring-primary/30 focus-visible:outline-none disabled:cursor-not-allowed disabled:bg-line-strong disabled:text-subtle disabled:shadow-none'

export const outlineButtonClass =
  'flex h-11 items-center justify-center gap-2 rounded-control border border-line-strong bg-white px-4 text-sm font-medium text-ink transition hover:bg-surface-2 focus-visible:ring-3 focus-visible:ring-chart/20 focus-visible:outline-none disabled:opacity-60'
