import { Condition } from '/@/lib/apis'
import ConditionDetail from './ConditionItem'
import { useEffect } from 'react'
import ReactTooltip from 'react-tooltip'

interface Props {
  conditions: Condition[]
}

const ConditionList = ({ conditions }: Props) => {
  useEffect(() => {
    ReactTooltip.rebuild()
  })
  return (
    <div className="border-primary min-h-64 flex flex-col items-center w-full border-b border-t">
      {conditions.map((condition, i) => (
        <ConditionDetail key={i} condition={condition} />
      ))}
    </div>
  )
}

export default ConditionList
