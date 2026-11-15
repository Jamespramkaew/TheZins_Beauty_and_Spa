import { useCallback, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { FiClock, FiPlus, FiRepeat } from 'react-icons/fi'
import { branchApi } from '../../api/branchApi'
import { stockApi } from '../../api/stockApi'
import { useAuth } from '../../hooks/useAuth'
import ContainersDrawer from './components/ContainersDrawer'
import DeleteStockDialog from './components/DeleteStockDialog'
import LogsTab from './components/LogsTab'
import MovementModal from './components/MovementModal'
import ProductFormModal from './components/ProductFormModal'
import ProductsTab from './components/ProductsTab'
import QuantitiesTab from './components/QuantitiesTab'
import TransferModal from './components/TransferModal'
import { outlineButtonClass, primaryButtonClass } from './components/styles'
import { DEFAULT_LOG_FILTERS, DEFAULT_QTY_FILTERS, formatDateTime } from './stockUtils'
import { useQuery } from '../../hooks/useQuery'

const TABS = [
  { value: 'quantities', label: 'ยอดคงเหลือ' },
  { value: 'logs', label: 'ประวัติเคลื่อนไหว' },
  { value: 'products', label: 'รายการสินค้า' },
]

// RBAC: the backend locks branch accounts to their own branch on every stock-log / container / employee endpoint
// and only admin may write products. This page mirrors that: branch accounts get no branch pickers (a locked
// chip instead), a read-only product list, and a transfer preview without the destination's stock.
function StocksPage() {
  const { user } = useAuth()
  const isAdmin = user.role === 'admin'
  const [searchParams, setSearchParams] = useSearchParams()
  const tab = TABS.some((item) => item.value === searchParams.get('tab')) ? searchParams.get('tab') : 'quantities'

  const [refreshKey, setRefreshKey] = useState(0)
  const [lastUpdated, setLastUpdated] = useState(null)
  const [qtyFilters, setQtyFilters] = useState(DEFAULT_QTY_FILTERS)
  const [logFilters, setLogFilters] = useState(DEFAULT_LOG_FILTERS)
  const [modal, setModal] = useState(null)

  const branchQuery = useQuery(branchApi.getAll, {})
  const stocks = useQuery(stockApi.getAll, {}, { refreshKey })
  const branches = branchQuery.data ?? []
  const stockList = stocks.data ?? []
  const ownBranch = isAdmin
    ? null
    : { id: user.branch_id, name: branches.find((branch) => branch.id === user.branch_id)?.name ?? 'สาขาของฉัน' }
  const activeStockCount = stocks.data ? stockList.filter((stock) => stock.is_active).length : null
  // Forms need the branch list for their pickers and defaults
  const formsReady = branchQuery.loaded && stocks.loaded

  const setTab = (value) => setSearchParams(value === 'quantities' ? {} : { tab: value }, { replace: true })
  const refresh = () => setRefreshKey((key) => key + 1)
  const closeModal = () => setModal(null)
  const onLoaded = useCallback((at) => setLastUpdated(at), [])

  const onRowAction = (kind, row, type) => {
    if (kind === 'movement') setModal({ kind, preset: { type, stockId: row.stock_id, branchId: row.branch_id } })
    else if (kind === 'transfer') setModal({ kind, preset: { stockId: row.stock_id, branchId: row.branch_id } })
    else if (kind === 'containers') {
      setModal({
        kind,
        stock: { id: row.stock_id, name: row.stock_name, uses_per_unit: row.uses_per_unit },
        branchId: row.branch_id,
      })
    } else if (kind === 'history') {
      setLogFilters({ ...DEFAULT_LOG_FILTERS, preset: 'month', stockId: row.stock_id, branchId: isAdmin ? row.branch_id : '' })
      setTab('logs')
    }
  }

  const afterSave = () => {
    closeModal()
    refresh()
  }

  const formProps = { isAdmin, branches, ownBranch, stocks: stockList, onClose: closeModal, onSaved: afterSave }

  return (
    <div className="mx-auto max-w-[1400px] space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-[26px] font-bold tracking-tight text-ink">สต็อกสินค้า</h1>
          <p className="mt-1 flex items-center gap-1.5 text-[13px] text-muted">
            <FiClock className="size-3.5" />
            {lastUpdated ? `อัปเดตล่าสุด ${formatDateTime(lastUpdated)} น.` : 'กำลังโหลดข้อมูล…'}
          </p>
        </div>
        <div className="flex gap-2.5">
          <button
            type="button"
            disabled={!formsReady}
            onClick={() => setModal({ kind: 'transfer', preset: null })}
            className={outlineButtonClass}
          >
            <FiRepeat className="size-4" />
            โอนสาขา
          </button>
          <button
            type="button"
            disabled={!formsReady}
            onClick={() => setModal({ kind: 'movement', preset: null })}
            className={primaryButtonClass}
          >
            <FiPlus className="size-4" />
            บันทึกรายการ
          </button>
        </div>
      </div>

      <div role="tablist" aria-label="ส่วนของหน้าสต็อก" className="flex gap-1 overflow-x-auto border-b border-line">
        {TABS.map((item) => {
          const selected = item.value === tab
          return (
            <button
              key={item.value}
              type="button"
              role="tab"
              id={`stocks-tab-${item.value}`}
              aria-selected={selected}
              aria-controls={`stocks-panel-${item.value}`}
              onClick={() => setTab(item.value)}
              className={`-mb-px border-b-2 px-3.5 py-3 text-sm whitespace-nowrap transition focus-visible:ring-3 focus-visible:ring-chart/20 focus-visible:outline-none ${
                selected ? 'border-primary font-semibold text-ink' : 'border-transparent text-muted hover:text-ink'
              }`}
            >
              {item.label}
            </button>
          )
        })}
      </div>

      <div role="tabpanel" id={`stocks-panel-${tab}`} aria-labelledby={`stocks-tab-${tab}`}>
        {tab === 'quantities' && (
          <QuantitiesTab
            isAdmin={isAdmin}
            branches={branches}
            ownBranch={ownBranch}
            activeStockCount={activeStockCount}
            filters={qtyFilters}
            onFiltersChange={setQtyFilters}
            refreshKey={refreshKey}
            onLoaded={onLoaded}
            onAction={onRowAction}
          />
        )}
        {tab === 'logs' && (
          <LogsTab
            isAdmin={isAdmin}
            branches={branches}
            stocks={stockList}
            filters={logFilters}
            onFiltersChange={setLogFilters}
            refreshKey={refreshKey}
            onLoaded={onLoaded}
          />
        )}
        {tab === 'products' && (
          <ProductsTab
            isAdmin={isAdmin}
            stocks={stocks}
            onCreate={() => setModal({ kind: 'product', stock: null })}
            onEdit={(stock) => setModal({ kind: 'product', stock })}
            onDelete={(stock) => setModal({ kind: 'delete', stock })}
          />
        )}
      </div>

      {modal?.kind === 'movement' && <MovementModal {...formProps} preset={modal.preset} />}
      {modal?.kind === 'transfer' && <TransferModal {...formProps} preset={modal.preset} />}
      {modal?.kind === 'containers' && (
        <ContainersDrawer
          isAdmin={isAdmin}
          branches={branches}
          stock={modal.stock}
          branchId={modal.branchId}
          onClose={closeModal}
        />
      )}
      {isAdmin && modal?.kind === 'product' && (
        <ProductFormModal stock={modal.stock} onClose={closeModal} onSaved={afterSave} />
      )}
      {isAdmin && modal?.kind === 'delete' && (
        <DeleteStockDialog stock={modal.stock} onClose={closeModal} onDeleted={afterSave} onDeactivated={afterSave} />
      )}
    </div>
  )
}

export default StocksPage
