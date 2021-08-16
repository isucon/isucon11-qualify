import { Condition } from '/@/lib/apis'
import ConditionDetail from './ConditionItem'

interface Props {
  conditions: Condition[]
}

const ConditionList = ({ conditions }: Props) => {
  return (
    <div className="border-primary flex flex-col items-center w-full border-b border-t">
      {conditions.map((condition, i) => (
        <ConditionDetail key={i} condition={condition} />
      ))}
    </div>
  )
}

export default ConditionList
