import Card from '../../../components/Card'
import CardLink from './CardLink'

function TopCoursesCard({ courses }) {
  const max = Math.max(...courses.map((course) => course.sold))

  return (
    <Card title="คอร์สขายดี" aside="จำนวนที่ขาย">
      <ol className="space-y-4">
        {courses.map((course, index) => (
          <li key={course.id} className="flex items-center gap-3">
            <span className="flex size-6 shrink-0 items-center justify-center rounded-md bg-surface-2 text-xs font-semibold text-muted">
              {index + 1}
            </span>
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium text-ink">{course.name}</p>
              <div className="mt-2 h-1 rounded-full bg-[#F0F0F3]">
                <div
                  className={`h-full rounded-full ${index === 0 ? 'bg-gold' : 'bg-[#C9CBD3]'}`}
                  style={{ width: `${(course.sold / max) * 100}%` }}
                />
              </div>
            </div>
            <span className="w-8 text-right text-sm font-semibold text-ink tabular-nums">{course.sold}</span>
          </li>
        ))}
      </ol>
      <CardLink to="/courses">ดูคอร์สทั้งหมด</CardLink>
    </Card>
  )
}

export default TopCoursesCard
