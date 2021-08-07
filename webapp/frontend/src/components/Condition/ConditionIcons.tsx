import ConditionIcon from './ConditionIcon'

interface Props {
  conditionCSV: string
}

const ConditionIcons = ({ conditionCSV }: Props) => {
  // 後でちゃんとパースする
  const is_dirty = true
  const is_overweight = true
  const is_broken = false

  return (
    <div>
      <ConditionIcon name={'is_dirty'} status={is_dirty} />
      <ConditionIcon name={'is_overweight'} status={is_overweight} />
      <ConditionIcon name={'is_broken'} status={is_broken} />
    </div>
  )
}

export default ConditionIcons
