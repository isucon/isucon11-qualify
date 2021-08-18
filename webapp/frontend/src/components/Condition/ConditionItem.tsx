import { Condition } from '/@/lib/apis'
import { getConditionTime } from '/@/lib/date'
import Tip from '/@/components/UI/Tip'
import ConditionIcons from './ConditionIcons'
import IsuStatusIcon from './IsuStatusIcon'

interface Props {
  condition: Condition
}

const ConditionItem = ({ condition }: Props) => {
  return (
    <div className="relative px-4 w-full">
      <div className="grid-cols-[min-content,1fr,min-content,5rem,6rem] grid gap-4 items-center">
        <div className="flex flex-wrap justify-center h-full">
          <div className="flex items-center h-full">
            <IsuStatusIcon
              condition_level={condition.condition_level}
              size={28}
            />
          </div>
          <div className="w-1px bg-line -translate-x-0.5px h-full transform -translate-y-full" />
        </div>
        <div className="text-primary break-words overflow-hidden">
          {condition.message}
        </div>
        <ConditionIcons conditionCSV={condition.condition} />
        {condition.is_sitting ? <Tip variant="sitting" /> : <div />}
        <div className="text-secondary my-4 text-center">
          {getConditionTime(condition.date)}
        </div>
      </div>
    </div>
  )
}

export default ConditionItem
