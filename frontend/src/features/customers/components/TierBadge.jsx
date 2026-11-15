import { FiStar } from 'react-icons/fi'
import { tierOf } from '../customerUtils'

function TierBadge({ tier }) {
  const { label, className } = tierOf(tier)

  return (
    <span className={`inline-flex items-center gap-1 rounded-md px-2 py-0.5 text-xs font-medium ${className}`}>
      {tier === 'vip' && <FiStar className="size-3" aria-hidden="true" />}
      {label}
    </span>
  )
}

export default TierBadge
