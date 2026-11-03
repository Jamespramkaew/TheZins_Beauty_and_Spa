import Toggle from './Toggle'

// Admin: switch (PATCH is_active) · branch account: read-only pill
function CourseStatus({ course, canManage, pending, onToggle }) {
  const label = course.is_active ? 'เปิดขาย' : 'ปิดใช้งาน'

  if (!canManage) {
    return (
      <span
        className={`inline-flex rounded-full px-2.5 py-1 text-xs font-medium ${
          course.is_active ? 'bg-success-bg text-success' : 'bg-surface-2 text-muted'
        }`}
      >
        {label}
      </span>
    )
  }

  return (
    <Toggle checked={course.is_active} onChange={onToggle} disabled={pending} label={`เปิดขายคอร์ส ${course.name}`}>
      <span className="whitespace-nowrap">{label}</span>
    </Toggle>
  )
}

export default CourseStatus
