import { useEffect } from 'react'
import { useState } from 'react'
import apis, { TrendResponse, Trend } from '/@/lib/apis'
import NowLoading from '/@/components/UI/NowLoading'
import TrendElement from './Trend'
import TrendHeadeer from './TrendHeader'
import ReactTooltip from 'react-tooltip'

const calcAllConditionLength = (trend: Trend) => {
  return trend.info.length + trend.warning.length + trend.critical.length
}

const TrendList = () => {
  const [trends, setTrends] = useState<TrendResponse>([])
  const [maxConditionCount, setMaxConditionCount] = useState(0)

  useEffect(() => {
    const update = async () => {
      const newTrends = await apis.getTrend()
      newTrends.sort(
        (a, b) => calcAllConditionLength(b) - calcAllConditionLength(a)
      )
      setTrends(newTrends)

      let max = 0
      newTrends.forEach(v => {
        const tmpLen = calcAllConditionLength(v)
        if (tmpLen > max) {
          max = tmpLen
        }
      })
      setMaxConditionCount(max)
    }
    update()
  }, [])
  useEffect(() => {
    ReactTooltip.rebuild()
  })

  return (
    <div className="relative">
      <h2 className="mb-8 text-2xl font-bold">みんなのISU</h2>
      <div className="mb-2">
        <TrendHeadeer />
      </div>
      {trends.map(trend => (
        <TrendElement
          key={trend.character}
          trend={trend}
          maxConditionCount={maxConditionCount}
        />
      ))}
      {trends.length === 0 ? <NowLoading /> : null}
    </div>
  )
}

export default TrendList
