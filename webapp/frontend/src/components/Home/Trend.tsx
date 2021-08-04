import { Trend } from '../../lib/apis'

interface Props {
  trend: Trend
}

const TrendElement = ({ trend }: Props) => {
  return (
    <div>
      <div>{trend.character}</div>
    </div>
  )
}

export default TrendElement
