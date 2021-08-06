import { TrendCondition } from '../../lib/apis'

interface Props {
  conditions: TrendCondition[]
  maxConditionCount: number
}

const calcWidthPercentage = (
  conditionLength: number,
  maxConditionCount: number
) => {
  return `${Math.round((conditionLength / maxConditionCount) * 100)}%`
}

const divideConditions = (conditions: TrendCondition[]) => {
  const info = conditions.filter(v => v.condition_level === 'info')
  const warning = conditions.filter(v => v.condition_level === 'warning')
  const critical = conditions.filter(v => v.condition_level === 'critical')
  return { info, warning, critical }
}

const TrendGraph = ({ conditions, maxConditionCount }: Props) => {
  const divided = divideConditions(conditions)
  return (
    <div className="flex items-center w-full h-4">
      <div
        className="h-full bg-status-info"
        style={{
          width: calcWidthPercentage(divided.info.length, maxConditionCount)
        }}
      />
      <div
        className="h-full bg-status-warning"
        style={{
          width: calcWidthPercentage(divided.warning.length, maxConditionCount)
        }}
      />
      <div
        className="h-full bg-status-critical"
        style={{
          width: calcWidthPercentage(divided.critical.length, maxConditionCount)
        }}
      />
    </div>
  )
}

export default TrendGraph
