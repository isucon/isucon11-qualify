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

    const status = sprited[1] === 'true' ? true : false
    statusPairs.push([sprited[0], status])
  })

  return (
    <div>
      {statusPairs.map(pair => (
        <div key={pair[0]}>
          <ConditionIcon name={pair[0]} status={pair[1]} />
        </div>
      ))}
    </div>
  )
}

export default ConditionIcons
