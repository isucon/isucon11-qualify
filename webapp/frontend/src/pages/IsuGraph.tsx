import { useEffect } from 'react'
import { useState } from 'react'
import Card from '../components/UI/Card'
import NowLoading from '../components/UI/NowLoading'
import TransitionGraph from '../components/IsuGraph/TransitionGraph'
import SittingGraph from '../components/IsuGraph/SittingGraph'
import Score from '../components/IsuGraph/Score'
import apis, { Graph, Isu } from '../lib/apis'

interface Props {
  isu: Isu
  setIsu: React.Dispatch<React.SetStateAction<Isu | null>>
}

const IsuGraph = ({ isu, setIsu }: Props) => {
  const [isuGraphs, setIsuGraphs] = useState<Graph[] | null>(null)
  const id = isu.jia_isu_uuid
  useEffect(() => {
    const load = async () => {
      // TODO: dateの取得方法を直す
      setIsuGraphs(await apis.getIsuGraphs(id, '2021-06-16%2B07:00'))
    }
    load()
  }, [id])

  if (!isu || !isuGraphs) {
    return <NowLoading />
  }
  return (
    <div>
      <Card>
        <div>
          <TransitionGraph isuGraphs={isuGraphs} />
          <SittingGraph isuGraphs={isuGraphs} />
          <Score isuGraphs={isuGraphs} />
        </div>
      </Card>
    </div>
  )
}

export default IsuGraph
