function Card({ title, aside, className = '', children }) {
  return (
    <section className={`flex flex-col rounded-card border border-line bg-white p-5 ${className}`}>
      {(title || aside) && (
        <header className="mb-4 flex items-start justify-between gap-3">
          <h2 className="text-[15px] font-semibold text-ink">{title}</h2>
          {aside && <div className="shrink-0 text-xs text-muted">{aside}</div>}
        </header>
      )}
      {children}
    </section>
  )
}

export default Card
