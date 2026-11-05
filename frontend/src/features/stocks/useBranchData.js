import { useMemo } from 'react'
import { employeeApi } from '../../api/employeeApi'
import { stockLogApi } from '../../api/stockLogApi'
import { fetchAllPages } from './stockUtils'
import { useQuery } from '../../hooks/useQuery'

const fetchBranchQuantities = (params) => fetchAllPages(stockLogApi.getQuantities, params)

const statusOf = (query) => (query.error ? 'error' : query.loaded && !query.loading ? 'ready' : 'loading')

// Active employees of one branch. Branch accounts can't pick: the backend always answers with their own branch.
export function useEmployees(branchId, isAdmin) {
  const query = useQuery(employeeApi.getAll, isAdmin ? { branch_id: branchId } : {}, { enabled: Boolean(branchId) })
  return { employees: branchId ? (query.data ?? []) : [], status: branchId ? statusOf(query) : 'idle' }
}

// On-hand quantity of every product at one branch, keyed by stock id.
// Branch accounts may only read their own branch, so pass enabled=false for any other branch.
export function useBranchQuantities(branchId, isAdmin, { enabled = true, refreshKey = 0 } = {}) {
  const query = useQuery(fetchBranchQuantities, isAdmin ? { branch_id: branchId, active: 'all' } : { active: 'all' }, {
    enabled: enabled && Boolean(branchId),
    refreshKey,
  })
  const byStock = useMemo(() => new Map((query.data ?? []).map((row) => [row.stock_id, row])), [query.data])
  return { byStock, status: statusOf(query) }
}

// Picker options: non-deleted products (GET /stocks), active ones only — but keep a preselected inactive one
export function productOptions(stocks, byStock, keepId) {
  return stocks
    .filter((stock) => stock.is_active || stock.id === keepId)
    .map((stock) => {
      const row = byStock.get(stock.id)
      return { id: stock.id, name: stock.name, quantity: row ? row.quantity : null, minQty: stock.min_qty }
    })
}
