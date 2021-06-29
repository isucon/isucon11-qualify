import { Condition } from '../../lib/apis'
import Tip from './Tip'

interface Props {
  condition: Condition
}

const getTime = (condition: Condition) => {
  const date = new Date(condition.timestamp)
  // 2020/01/01 01:01:01
  return `${date.getFullYear()}/${pad0(date.getMonth() + 1)}/${pad0(
    date.getDay()
  )} ${pad0(date.getHours())}:${pad0(date.getMinutes())}:${pad0(
    date.getSeconds()
  )}`
}

const pad0 = (num: number) => ('0' + num).slice(-2)

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
        <div>{getTime(condition)}</div>
      </div>
    </div>
  )
}

export default ConditionDetail
