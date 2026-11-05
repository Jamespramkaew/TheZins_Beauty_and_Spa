import dayjs from 'dayjs'
import buddhistEra from 'dayjs/plugin/buddhistEra'
import 'dayjs/locale/th'

dayjs.extend(buddhistEra)

// Thai dates with the Buddhist year, e.g. "6 พ.ย. 2569"
export const formatDate = (value) => (value ? dayjs(value).locale('th').format('D MMM BBBB') : '–')

export const formatDateTime = (value) => (value ? dayjs(value).locale('th').format('D MMM BBBB, HH:mm') : '–')

// Date-only columns (b_day, order_date) come back at UTC midnight — keep the YYYY-MM-DD part
// so no timezone can shift the day
export const formatDateOnly = (value) => (value ? formatDate(value.slice(0, 10)) : '–')

// YYYY-MM-DD of the browser's local day, for <input type="date"> and ?from= / ?to=
export const toISODate = (value) => dayjs(value).format('YYYY-MM-DD')
