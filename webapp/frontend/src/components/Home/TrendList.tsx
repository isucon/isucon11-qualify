import { useEffect } from 'react'
import { useState } from 'react'
import apis, { TrendResponse } from '../../lib/apis'
import NowLoading from '../UI/NowLoading'
import TrendElement from './Trend'

const TrendList = () => {
  const [trends, setTrends] = useState<TrendResponse>([])
  useEffect(() => {
    const update = async () => {
      setTrends(await apis.getTrend())
    }
    update()
  }, [])

  if (trends.length === 0) {
    return <NowLoading />
  }
  return (
    <div>
      {trends.map(trend => (
        <TrendElement key={trend.character} trend={trend} />
      ))}
    </div>
  )
}

export default TrendList
