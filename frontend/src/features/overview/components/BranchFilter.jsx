import { useEffect, useRef, useState } from 'react'
import { LuCheck, LuStore, LuX } from 'react-icons/lu'

function BranchFilter({ branches, selectedIds, onChange }) {
  const [open, setOpen] = useState(false)
  const containerRef = useRef(null)
  const selected = branches.filter((branch) => selectedIds.includes(branch.id))
  const allSelected = selected.length === branches.length

  useEffect(() => {
    if (!open) return
    const onPointerDown = (event) => {
      if (!containerRef.current?.contains(event.target)) setOpen(false)
    }
    const onKeyDown = (event) => event.key === 'Escape' && setOpen(false)
    document.addEventListener('pointerdown', onPointerDown)
    document.addEventListener('keydown', onKeyDown)
    return () => {
      document.removeEventListener('pointerdown', onPointerDown)
      document.removeEventListener('keydown', onKeyDown)
    }
  }, [open])

  const toggle = (id) => {
    const next = selectedIds.includes(id)
      ? selectedIds.filter((selectedId) => selectedId !== id)
      : [...selectedIds, id]
    // Keep at least one branch selected
    if (next.length) onChange(next)
  }

  return (
    <div ref={containerRef} className="relative inline-flex">
      <div className="flex h-9 items-center rounded-control border border-line-strong bg-white text-sm">
        <button
          type="button"
          onClick={() => setOpen((value) => !value)}
          aria-expanded={open}
          className="flex h-full items-center gap-2 pr-2 pl-3"
        >
          <LuStore className="size-4 text-muted" />
          <span className="font-medium text-ink">สาขา</span>
          <span className="text-xs text-subtle">is any of</span>
          <span className="flex items-center gap-1.5 rounded-md bg-surface-2 px-2 py-0.5 text-xs font-medium text-ink">
            <span className="flex -space-x-1">
              {selected.map((branch) => (
                <span
                  key={branch.id}
                  className="size-2.5 rounded-full ring-2 ring-surface-2"
                  style={{ backgroundColor: branch.color }}
                />
              ))}
            </span>
            {allSelected ? `${branches.length} สาขา` : selected.map((branch) => branch.name).join(', ')}
          </span>
        </button>
        {!allSelected && (
          <button
            type="button"
            onClick={() => onChange(branches.map((branch) => branch.id))}
            aria-label="ล้างตัวกรองสาขา"
            className="mr-1 flex size-7 items-center justify-center rounded-md text-subtle hover:bg-surface-2 hover:text-ink"
          >
            <LuX className="size-3.5" />
          </button>
        )}
      </div>

      {open && (
        <ul className="absolute top-full left-0 z-10 mt-2 w-56 rounded-control border border-line bg-white p-1.5 shadow-[0_12px_32px_-12px_rgba(23,24,28,0.25)]">
          {branches.map((branch) => {
            const checked = selectedIds.includes(branch.id)
            return (
              <li key={branch.id}>
                <button
                  type="button"
                  role="menuitemcheckbox"
                  aria-checked={checked}
                  onClick={() => toggle(branch.id)}
                  className="flex min-h-10 w-full items-center gap-2.5 rounded-lg px-2.5 text-sm text-ink hover:bg-surface-2"
                >
                  <span
                    className={`flex size-4 items-center justify-center rounded border ${
                      checked ? 'border-primary bg-primary text-white' : 'border-line-strong'
                    }`}
                  >
                    {checked && <LuCheck className="size-3" />}
                  </span>
                  <span className="size-2.5 rounded-sm" style={{ backgroundColor: branch.color }} />
                  {branch.name}
                </button>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}

export default BranchFilter
