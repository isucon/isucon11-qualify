import { Condition } from '../../lib/apis'

interface Props {
  condition: Condition
}

const ConditionDetail = ({ condition }: Props) => {
  return (
    <div>
      <div>{JSON.stringify(condition)}</div>
    </div>
  )
}

export default ConditionDetail
