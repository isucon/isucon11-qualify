import { useEffect } from 'react'
import { useState } from 'react'
import apis, { TrendResponse } from '../../lib/apis'
import NowLoading from '../UI/NowLoading'
import TrendElement from './Trend'
import TrendHeadeer from './TrendHeader'

const TrendList = () => {
  const [trends, setTrends] = useState<TrendResponse>([])
  const [maxConditionCount, setMaxConditionCount] = useState(0)
  useEffect(() => {
    const update = async () => {
      const newTrends = await apis.getTrend()
      newTrends.sort((a, b) => b.conditions.length - a.conditions.length)
      setTrends(newTrends)

      let max = 0
      newTrends.forEach(v => {
        if (v.conditions.length > max) {
          max = v.conditions.length
        }
      })
      setMaxConditionCount(max)
    }
    update()
  }, [])

  if (trends.length === 0) {
    return <NowLoading />
  }
  return (
    <div>
      <h2 className="mb-6 text-xl font-bold">みんなのISU</h2>
      <TrendHeadeer />
      {trends.map(trend => (
        <TrendElement
          key={trend.character}
          trend={trend}
          maxConditionCount={maxConditionCount}
        />
      ))}
    </div>
  )
}

export default TrendList
