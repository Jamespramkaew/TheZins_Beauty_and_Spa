// Order comes from the API — never re-sort here. `limit` collapses the rest into a "+N" chip.
function ItemChips({ items, limit }) {
  if (!items.length) return <p className="text-[13px] text-subtle italic">ไม่ใช้สินค้า</p>

  const shown = limit ? items.slice(0, limit) : items
  const hidden = items.slice(shown.length)

  return (
    <ul className="flex flex-wrap gap-1.5">
      {shown.map((item) => (
        <li
          key={item.stock.id}
          className="max-w-full rounded-md border border-line bg-surface-2 px-2 py-0.5 text-xs text-ink [overflow-wrap:anywhere]"
        >
          {item.stock.name} ×{item.use_qty}
        </li>
      ))}
      {hidden.length > 0 && (
        <li
          title={hidden.map((item) => `${item.stock.name} ×${item.use_qty}`).join(', ')}
          className="rounded-md bg-gold-bg px-2 py-0.5 text-xs font-semibold text-gold-ink"
        >
          +{hidden.length}
        </li>
      )}
    </ul>
  )
}

export default ItemChips
