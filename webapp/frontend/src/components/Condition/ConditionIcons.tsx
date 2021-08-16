import ConditionIcon from './ConditionIcon'

interface Props {
  conditionCSV: string
}

const ConditionIcons = ({ conditionCSV }: Props) => {
  const list = conditionCSV.split(',')
  const statusPairs: [string, boolean][] = []
  list.forEach(element => {
    const sprited = element.split('=')
    if (sprited.length !== 2) return

    const status = sprited[1] === 'true'
    statusPairs.push([sprited[0], status])
  })

  return (
    <div className="flex gap-3 items-center">
      {statusPairs.map(pair => (
        <ConditionIcon key={pair[0]} name={pair[0]} status={pair[1]} />
      ))}
    </div>
  )
}

export default ConditionIcons
