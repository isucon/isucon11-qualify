import { Trend } from '/@/lib/apis'
import { getConditionTime } from '/@/lib/date'
import TrendGraph from './TrendGraph'

interface Props {
  trend: Trend
  maxConditionCount: number
}

const getLatestStringTime = (trend: Trend) => {
  let latest: Date | null = null
  for (const c of [trend.info, trend.warning, trend.critical].flat()) {
    if (!latest || latest < c.date) {
      latest = c.date
    }
  }
  if (!latest) {
    return 'no data'
  }
  return getConditionTime(latest)
}

const TrendElement = ({ trend, maxConditionCount }: Props) => {
  return (
    <div className="grid-cols-trend grid p-2">
      <div className="flex flex-col">
        <div>{trend.character}</div>
        <div className="text-secondary">{getLatestStringTime(trend)}</div>
      </div>
      <div className="flex items-center">
        <TrendGraph trend={trend} maxConditionCount={maxConditionCount} />
      </div>
    </div>
  )
}

export default TrendElement
