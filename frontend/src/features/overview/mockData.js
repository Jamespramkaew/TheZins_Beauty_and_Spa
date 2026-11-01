// Mock data for the Overview page. The backend has no overview endpoint yet
// (see ui-handoff.md §8.3) — replace with API data once it exists.

// Branch colors are assigned by branch order and must not repeat
export const BRANCH_COLORS = ['#2F4C8F', '#8C2F3D', '#A7B6C6']

export const branches = [
  { id: 1, name: 'รังสิต', sales: 318500 },
  { id: 2, name: 'บางนา', sales: 251200 },
  { id: 3, name: 'ลาดพร้าว', sales: 172600 },
].map((branch, index) => ({ ...branch, color: BRANCH_COLORS[index] }))

export const periodOptions = [
  { value: 'today', label: 'วันนี้' },
  { value: 'week', label: 'สัปดาห์นี้' },
  { value: 'month', label: 'เดือนนี้' },
  { value: 'custom', label: 'กำหนดเอง' },
]

export const currentMonthLabel = 'ก.ย. 2569'
export const previousMonthLabel = 'ส.ค. 2569'

const currentMonthSales = [
  22400, 21800, 19600, 19100, 26300, 29200, 28400, 23400, 23200, 24100,
  25600, 27400, 25800, 31200, 26700, 24800, 24300, 23300, 22800, 26100,
  24900, 23700, 22600, 21900, 23500, 25200, 24400, 22100, 19800, 38700,
]

const previousMonthSales = [
  20200, 18900, 19800, 17100, 23700, 23200, 21800, 24600, 22400, 22900,
  21300, 22600, 23800, 24400, 19900, 22100, 20600, 19700, 21400, 22500,
  21900, 23300, 22800, 21700, 20900, 22200, 23400, 19500, 18200, 35400,
]

export const dailySales = currentMonthSales.map((sales, index) => ({
  day: index + 1,
  sales,
  previous: previousMonthSales[index],
}))

export const kpis = {
  todayQueues: 18,
  newCustomers: 24,
}

export const topCourses = [
  { id: 1, name: 'หน้าใส Aura', sold: 32 },
  { id: 2, name: 'เลเซอร์กำจัดขน', sold: 21 },
  { id: 3, name: 'ทรีตเมนต์ลดสิว', sold: 17 },
]

export const topEmployee = {
  name: 'คุณพิมพ์ชนก ทองดี',
  initials: 'พพ',
  branchId: 1,
  sales: 128600,
}

export const stockAlerts = [
  { id: 1, name: 'ครีมกันแดด SPF50', note: 'ใกล้หมด', remaining: 3, icon: 'sun' },
  { id: 2, name: 'มาส์กหน้าไฮยา', note: 'ต้องสั่งเพิ่ม', remaining: 0, icon: 'droplet' },
]
