import { Trend } from '../../lib/apis'

interface Props {
  trend: Trend
  maxConditionCount: number
}

const calcWidthPercentage = (
  conditionLength: number,
  maxConditionCount: number
) => {
  return `${Math.round((conditionLength / maxConditionCount) * 100)}%`
}

const TrendGraph = ({ trend, maxConditionCount }: Props) => {
  return (
    <div className="flex items-center w-full h-4">
      <div
        className="h-full bg-status-info"
        style={{
          width: calcWidthPercentage(trend.info.length, maxConditionCount)
        }}
      />
      <div
        className="h-full bg-status-warning"
        style={{
          width: calcWidthPercentage(trend.warning.length, maxConditionCount)
        }}
      />
      <div
        className="h-full bg-status-critical"
        style={{
          width: calcWidthPercentage(trend.critical.length, maxConditionCount)
        }}
      />
    </div>
  )
}

export default TrendGraph
