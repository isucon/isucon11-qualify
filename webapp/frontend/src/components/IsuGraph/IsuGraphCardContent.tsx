import apis, { Graph, Isu } from '../../lib/apis'
import { useEffect } from 'react'
import { useState } from 'react'
import NowLoading from '../UI/NowLoading'
import TransitionGraph from './TransitionGraph'
import SittingGraph from './SittingGraph'
import Score from './Score'
import DateInput from './DateInput'

interface Props {
  isu: Isu
}

const IsuGraphCardContent = ({ isu }: Props) => {
  const [isuGraphs, setIsuGraphs] = useState<Graph[] | null>(null)
  const [date, setDate] = useState(new Date())
  const id = isu.jia_isu_uuid

  const search = async (date: Date) => {
    setIsuGraphs(
      await apis.getIsuGraphs(id, { date: Math.floor(date.getTime() / 1000) })
    )
  }

  useEffect(() => {
    const load = async () => {
      // TODO: dateの取得方法を直す
      search(date)
    }
    load()
  })

  if (!isuGraphs) {
    return <NowLoading />
  }
  return (
    <div>
      <DateInput date={date} search={search} />
      <TransitionGraph isuGraphs={isuGraphs} />
      <SittingGraph isuGraphs={isuGraphs} />
      <Score isuGraphs={isuGraphs} />
    </div>
  )
}

export default IsuGraphCardContent
