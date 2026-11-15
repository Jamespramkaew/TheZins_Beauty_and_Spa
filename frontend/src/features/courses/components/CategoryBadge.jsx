import { categoryLabel, categoryTone } from '../courseUtils'

function CategoryBadge({ category }) {
  if (!category) return null

  return (
    <span className={`inline-flex rounded-md px-2 py-0.5 text-xs font-medium ${categoryTone(category)}`}>
      {categoryLabel(category)}
    </span>
  )
}

export default CategoryBadge
