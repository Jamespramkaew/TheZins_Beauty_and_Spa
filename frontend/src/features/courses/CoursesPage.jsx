import { useEffect, useState } from 'react'
import { FiEye, FiGrid, FiList, FiPlus, FiSearch } from 'react-icons/fi'
import { courseApi } from '../../api/courseApi'
import { useAuth } from '../../hooks/useAuth'
import { useToast } from '../../hooks/useToast'
import { formatNumber } from '../../utils/format'
import CourseCard from './components/CourseCard'
import CourseDetailModal from './components/CourseDetailModal'
import CourseFormModal from './components/CourseFormModal'
import CourseTable from './components/CourseTable'
import DeleteCourseDialog from './components/DeleteCourseDialog'
import Pagination from '../../components/ui/Pagination'
import { DEFAULT_CATEGORIES, PAGE_SIZE_OPTIONS } from './courseUtils'
import { useCourses } from './useCourses'

const VIEWS = [
  { value: 'table', label: 'ตาราง', icon: FiList },
  { value: 'card', label: 'การ์ด', icon: FiGrid },
]

function CoursesPage() {
  const { user } = useAuth()
  const canManage = user.role === 'admin'
  const toast = useToast()
  const [searchInput, setSearchInput] = useState('')
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(PAGE_SIZE_OPTIONS[0])
  const [view, setView] = useState('table')
  const [viewing, setViewing] = useState(null)
  const [form, setForm] = useState(null) // { course } — course null = create
  const [deleting, setDeleting] = useState(null)
  const [togglingIds, setTogglingIds] = useState(() => new Set())
  const { courses, meta, error, loading, loaded, stats, reload, replaceCourse } = useCourses({ search, page, pageSize })

  useEffect(() => {
    const timer = setTimeout(() => {
      setSearch(searchInput.trim())
      setPage(1)
    }, 300)
    return () => clearTimeout(timer)
  }, [searchInput])

  const categories = [...new Set([...DEFAULT_CATEGORIES, ...(stats?.categories ?? [])])]

  const toggleActive = async (course) => {
    setTogglingIds((ids) => new Set(ids).add(course.id))
    try {
      const response = await courseApi.update(course.id, { is_active: !course.is_active })
      replaceCourse(response.data)
      toast.success(response.data.is_active ? 'เปิดขายแล้ว' : 'ปิดขายแล้ว', `“${course.name}”`)
    } catch {
      toast.error('เปลี่ยนสถานะไม่สำเร็จ', 'กรุณาลองใหม่อีกครั้ง')
    } finally {
      setTogglingIds((ids) => {
        const next = new Set(ids)
        next.delete(course.id)
        return next
      })
    }
  }

  const listProps = {
    canManage,
    onEdit: (course) => setForm({ course }),
    onDelete: (course) => setDeleting(course),
    onToggleActive: toggleActive,
  }

  return (
    <div className="mx-auto max-w-[1400px] space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-[28px] font-bold tracking-tight text-ink">คอร์สบริการ</h1>
          <p className="mt-1 text-sm text-muted">
            แพ็กเกจที่ลูกค้าซื้อไว้ใช้หลายครั้ง — กำหนดราคา จำนวนครั้ง อายุคอร์ส และค่ามือต่อครั้ง
          </p>
        </div>
        {canManage ? (
          <button
            type="button"
            onClick={() => setForm({ course: null })}
            className="flex h-11 items-center gap-2 rounded-control bg-primary px-4 text-sm font-semibold text-white shadow-sm transition hover:bg-[#4d1621]"
          >
            <FiPlus className="size-4" />
            เพิ่มคอร์ส
          </button>
        ) : (
          <span className="flex items-center gap-2 rounded-full border border-line bg-surface-2 px-3.5 py-1.5 text-sm text-muted">
            <FiEye className="size-4" />
            ดูอย่างเดียว (บัญชีสาขา)
          </span>
        )}
      </div>

      <dl className="grid grid-cols-2 gap-4 xl:grid-cols-4">
        <StatTile label="คอร์สทั้งหมด" value={stats?.total} />
        <StatTile label="เปิดขายอยู่" value={stats?.active} tone="text-success" />
        <StatTile label="ปิดใช้งาน" value={stats?.inactive} />
        <StatTile label="ไม่หมดอายุ" value={stats?.noExpiry} />
      </dl>

      <div className="flex flex-wrap items-center justify-between gap-3">
        <label className="relative w-full max-w-sm">
          <FiSearch className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-subtle" />
          <input
            type="search"
            value={searchInput}
            onChange={(event) => setSearchInput(event.target.value)}
            placeholder="ค้นหาชื่อคอร์ส"
            aria-label="ค้นหาชื่อคอร์ส"
            className="h-10 w-full rounded-control border border-line-strong bg-white pr-3 pl-10 text-sm placeholder:text-subtle focus:border-chart/40 focus:ring-3 focus:ring-chart/10 focus:outline-none"
          />
        </label>

        <div role="tablist" aria-label="รูปแบบการแสดงผล" className="flex rounded-[12px] border border-line bg-surface-2 p-1">
          {VIEWS.map(({ value, label, icon: Icon }) => (
            <button
              key={value}
              type="button"
              role="tab"
              aria-selected={view === value}
              onClick={() => setView(value)}
              className={`flex h-9 items-center gap-2 rounded-lg px-3.5 text-sm transition ${
                view === value ? 'bg-white font-semibold text-ink shadow-sm' : 'text-muted hover:text-ink'
              }`}
            >
              <Icon className="size-4" />
              {label}
            </button>
          ))}
        </div>
      </div>

      <section aria-busy={loading} className={`transition-opacity ${loading && loaded ? 'opacity-60' : ''}`}>
        {error ? (
          <EmptyState title="โหลดรายการคอร์สไม่สำเร็จ" detail="ตรวจสอบการเชื่อมต่อหรือการเข้าสู่ระบบ แล้วลองอีกครั้ง">
            <button
              type="button"
              onClick={reload}
              className="mt-4 h-10 rounded-control border border-line-strong bg-white px-4 text-sm font-medium text-ink hover:bg-surface-2"
            >
              ลองอีกครั้ง
            </button>
          </EmptyState>
        ) : !loaded ? (
          <EmptyState title="กำลังโหลดรายการคอร์ส…" />
        ) : courses.length === 0 ? (
          <EmptyState
            title={search ? 'ไม่พบคอร์สที่ค้นหา' : 'ยังไม่มีคอร์ส'}
            detail={search ? `ไม่มีคอร์สที่ชื่อมีคำว่า “${search}”` : canManage ? 'กด “เพิ่มคอร์ส” เพื่อเริ่มต้น' : undefined}
          />
        ) : view === 'card' ? (
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {courses.map((course) => (
              <CourseCard
                key={course.id}
                course={course}
                canManage={canManage}
                toggling={togglingIds.has(course.id)}
                onEdit={() => listProps.onEdit(course)}
                onDelete={() => listProps.onDelete(course)}
                onToggleActive={() => toggleActive(course)}
              />
            ))}
          </div>
        ) : (
          <CourseTable courses={courses} togglingIds={togglingIds} onView={setViewing} {...listProps} />
        )}
      </section>

      {viewing && (
        <CourseDetailModal
          course={viewing}
          canManage={canManage}
          onClose={() => setViewing(null)}
          onEdit={() => {
            setViewing(null)
            setForm({ course: viewing })
          }}
        />
      )}

      {meta && meta.total > 0 && !error && (
        <Pagination
          meta={meta}
          unit="คอร์ส"
          pageSizeOptions={PAGE_SIZE_OPTIONS}
          onPageChange={setPage}
          onPageSizeChange={(size) => {
            setPageSize(size)
            setPage(1)
          }}
        />
      )}

      {form && (
        <CourseFormModal
          course={form.course}
          categories={categories}
          onClose={() => setForm(null)}
          onSaved={() => {
            setForm(null)
            reload()
          }}
        />
      )}

      {deleting && (
        <DeleteCourseDialog
          course={deleting}
          onClose={() => setDeleting(null)}
          onDeactivated={(updated) => {
            setDeleting(null)
            replaceCourse(updated)
          }}
          onDeleted={() => {
            setDeleting(null)
            // Last card on a later page → step back a page
            if (courses.length === 1 && page > 1) setPage(page - 1)
            reload()
          }}
        />
      )}
    </div>
  )
}

function StatTile({ label, value, tone = 'text-ink' }) {
  return (
    <div className="rounded-card border border-line bg-white p-5">
      <dt className="text-sm text-muted">{label}</dt>
      <dd className="mt-3 flex items-baseline gap-1.5">
        <span className={`text-[26px] leading-none font-bold tracking-tight ${tone}`}>
          {value == null ? '–' : formatNumber(value)}
        </span>
        <span className="text-sm text-muted">คอร์ส</span>
      </dd>
    </div>
  )
}

function EmptyState({ title, detail, children }) {
  return (
    <div className="flex flex-col items-center justify-center rounded-card border border-dashed border-line-strong bg-surface px-6 py-16 text-center">
      <p className="font-semibold text-ink">{title}</p>
      {detail && <p className="mt-1 text-sm text-muted">{detail}</p>}
      {children}
    </div>
  )
}

export default CoursesPage
