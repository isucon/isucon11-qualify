import { Condition } from '/@/lib/apis'
import ConditionDetail from './ConditionDetail'

interface Props {
  conditions: Condition[]
}

const ConditionList = ({ conditions }: Props) => {
  return (
    <div className="flex flex-col gap-4 items-center w-full">
      <div className="w-full border border-b-0 border-outline">
        {conditions.map((condition, i) => (
          <div className="border-b border-outline" key={i}>
            <ConditionDetail condition={condition} />
          </div>
        ))}
      </div>
    </div>
  )
}

export default ConditionList
