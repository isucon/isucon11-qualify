import { Condition } from '../../lib/apis'
import { getConditionTime } from '../../lib/date'
import Tip from '../UI/Tip'

interface Props {
  condition: Condition
}

const ConditionDetail = ({ condition }: Props) => {
  return (
    <div className="flex flex-wrap gap-4 items-center p-4">
      <div className="mr-auto">
        <div>{condition.isu_name}</div>
        <div className="text-secondary">{condition.message}</div>
      </div>
      <div className="flex justify-center w-24">
        {condition.is_sitting ? <Tip variant="sitting" /> : null}
      </div>
      <div className="flex justify-center w-24">
        <Tip variant={condition.condition_level} />
      </div>
      <div>
        <div>{getConditionTime(condition)}</div>
      </div>
    </div>
  )
}

export default ConditionDetail
