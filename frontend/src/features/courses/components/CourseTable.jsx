import { FiEdit2, FiTrash2 } from 'react-icons/fi'
import { formatBaht, formatNumber } from '../../../utils/format'
import { formatValidDay, pricePerSession, toNumber } from '../courseUtils'
import CategoryBadge from './CategoryBadge'
import CourseStatus from './CourseStatus'
import ItemChips from './ItemChips'

function CourseTable({ courses, canManage, togglingIds, onView, onEdit, onDelete, onToggleActive }) {
  return (
    <div className="overflow-x-auto rounded-card border border-line">
      <table className="w-full min-w-[1080px] text-sm">
        <thead className="bg-surface text-xs text-muted">
          <tr className="text-left">
            <th className="px-4 py-3 font-medium">ชื่อคอร์ส</th>
            <th className="px-3 py-3 text-right font-medium">ราคา</th>
            <th className="px-3 py-3 text-right font-medium">จำนวนครั้ง</th>
            <th className="px-3 py-3 text-right font-medium">เฉลี่ย / ครั้ง</th>
            <th className="px-3 py-3 font-medium">อายุคอร์ส</th>
            <th className="px-3 py-3 text-right font-medium">ค่ามือ / ครั้ง</th>
            <th className="px-3 py-3 font-medium">สินค้าที่ใช้ / บริการ</th>
            <th className="px-3 py-3 font-medium">สถานะ</th>
            {canManage && <th className="px-4 py-3 text-right font-medium">การทำงาน</th>}
          </tr>
        </thead>
        <tbody className="divide-y divide-line">
          {courses.map((course) => (
            <tr key={course.id} className={course.is_active ? 'bg-white' : 'bg-surface text-muted'}>
              <td className="max-w-64 px-4 py-3.5">
                <button
                  type="button"
                  onClick={() => onView(course)}
                  title="ดูรายละเอียด"
                  className={`rounded-sm text-left font-semibold text-chart [overflow-wrap:anywhere] underline-offset-4 hover:underline focus-visible:ring-3 focus-visible:ring-chart/20 focus-visible:outline-none ${
                    course.is_active ? '' : 'opacity-70'
                  }`}
                >
                  {course.name}
                </button>
                <div className="mt-1.5">
                  <CategoryBadge category={course.category} />
                </div>
              </td>
              <td className="px-3 py-3.5 text-right font-semibold whitespace-nowrap">{formatBaht(toNumber(course.price))}</td>
              <td className="px-3 py-3.5 text-right whitespace-nowrap">{formatNumber(course.max_session)} ครั้ง</td>
              <td className="px-3 py-3.5 text-right whitespace-nowrap text-muted">
                {formatBaht(pricePerSession(course.price, course.max_session))}
              </td>
              <td className={`px-3 py-3.5 whitespace-nowrap ${course.valid_day == null ? 'text-chart' : ''}`}>
                {formatValidDay(course.valid_day)}
              </td>
              <td className="px-3 py-3.5 text-right whitespace-nowrap">{formatBaht(toNumber(course.service_comm))}</td>
              <td className="w-72 px-3 py-3.5">
                <ItemChips items={course.items} limit={2} onShowMore={() => onView(course)} />
              </td>
              <td className="px-3 py-3.5">
                <CourseStatus
                  course={course}
                  canManage={canManage}
                  pending={togglingIds.has(course.id)}
                  onToggle={() => onToggleActive(course)}
                />
              </td>
              {canManage && (
                <td className="px-4 py-3.5">
                  <div className="flex justify-end gap-2">
                    <IconButton label={`แก้ไขคอร์ส ${course.name}`} onClick={() => onEdit(course)}>
                      <FiEdit2 className="size-4" />
                    </IconButton>
                    <IconButton label={`ลบคอร์ส ${course.name}`} onClick={() => onDelete(course)} danger>
                      <FiTrash2 className="size-4" />
                    </IconButton>
                  </div>
                </td>
              )}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function IconButton({ label, onClick, danger, children }) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={label}
      className={`flex size-9 items-center justify-center rounded-control border border-line-strong bg-white transition ${
        danger ? 'text-danger hover:bg-danger-bg' : 'text-muted hover:bg-surface-2 hover:text-ink'
      }`}
    >
      {children}
    </button>
  )
}

export default CourseTable
