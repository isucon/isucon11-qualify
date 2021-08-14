import { Condition } from '/@/lib/apis'
import NowLoading from '/@/components/UI/NowLoading'
import ConditionDetail from './ConditionDetail'

interface Props {
  conditions: Condition[]
}

const Conditions = ({ conditions }: Props) => {
  if (!conditions) return <NowLoading />

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

export default Conditions
