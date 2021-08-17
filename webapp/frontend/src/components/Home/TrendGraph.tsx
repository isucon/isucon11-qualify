import { Trend } from '/@/lib/apis'

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
  const total = trend.info.length + trend.warning.length + trend.critical.length
  return (
    <div className="flex items-center w-full h-4">
      <div
        data-tip={`${trend.info.length}/${total}脚`}
        data-place="top"
        className="bg-status-info duration-50 hover:h-5 h-full"
        style={{
          width: calcWidthPercentage(trend.info.length, maxConditionCount)
        }}
      />
      <div
        data-tip={`${trend.warning.length}/${total}脚`}
        data-place="top"
        className="bg-status-warning duration-50 hover:h-5 h-full"
        style={{
          width: calcWidthPercentage(trend.warning.length, maxConditionCount)
        }}
      />
      <div
        data-tip={`${trend.critical.length}/${total}脚`}
        data-place="top"
        className="bg-status-critical duration-50 hover:h-5 h-full"
        style={{
          width: calcWidthPercentage(trend.critical.length, maxConditionCount)
        }}
      />
    </div>
  )
}

export default TrendGraph
