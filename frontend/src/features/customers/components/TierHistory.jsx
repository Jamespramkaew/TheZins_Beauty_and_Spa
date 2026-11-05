import { formatDateTime } from '../customerUtils'
import TierBadge from './TierBadge'

// rows = GET /customers/:id/tier-changes (newest first)
function TierHistory({ rows }) {
  if (rows.length === 0) {
    return (
      <p className="rounded-card border border-dashed border-line-strong bg-surface px-6 py-12 text-center text-sm text-muted">
        ยังไม่เคยเปลี่ยนระดับ — ลูกค้าใหม่เริ่มที่ระดับทั่วไป
      </p>
    )
  }

  return (
    <div className="overflow-x-auto rounded-card border border-line">
      <table className="w-full min-w-[520px] text-sm">
        <thead className="bg-surface text-xs text-muted">
          <tr className="text-left">
            <th className="px-4 py-3 font-medium">วันที่เปลี่ยน</th>
            <th className="px-3 py-3 font-medium">เปลี่ยนเป็น</th>
            <th className="px-4 py-3 font-medium">ผู้เปลี่ยน</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-line bg-white">
          {rows.map((row) => (
            <tr key={row.id}>
              <td className="px-4 py-3.5 whitespace-nowrap">{formatDateTime(row.changed_at)}</td>
              <td className="px-3 py-3.5">
                <TierBadge tier={row.new_tier} />
              </td>
              <td className="px-4 py-3.5 text-muted">{row.changed_by_username}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export default TierHistory
