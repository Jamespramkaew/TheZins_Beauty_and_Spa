import { Link } from 'react-router-dom'
import { LuArrowRight } from 'react-icons/lu'

function CardLink({ to, children }) {
  return (
    <Link
      to={to}
      className="mt-auto inline-flex items-center gap-1 self-start pt-5 text-sm font-medium text-ink underline decoration-line-strong underline-offset-4 transition hover:decoration-ink"
    >
      {children}
      <LuArrowRight className="size-3.5" />
    </Link>
  )
}

export default CardLink
