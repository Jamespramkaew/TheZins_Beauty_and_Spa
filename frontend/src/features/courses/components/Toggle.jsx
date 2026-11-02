function Toggle({ checked, onChange, disabled, label, children }) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      aria-label={label}
      disabled={disabled}
      onClick={() => onChange(!checked)}
      className="flex items-center gap-2.5 text-sm text-ink disabled:cursor-wait disabled:opacity-60"
    >
      <span className={`relative h-5.5 w-10 shrink-0 rounded-full transition ${checked ? 'bg-success' : 'bg-line-strong'}`}>
        <span
          className={`absolute top-0.5 left-0.5 size-4.5 rounded-full bg-white shadow-sm transition-transform ${
            checked ? 'translate-x-4.5' : ''
          }`}
        />
      </span>
      {children}
    </button>
  )
}

export default Toggle
