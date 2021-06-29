import { useEffect } from 'react'
import { useState } from 'react'
import Card from '../components/UI/Card'
import NowLoading from '../components/UI/NowLoading'
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
    console.log(isuGraphs)
  }, [id])

  if (!isu || !isuGraphs) {
    return <NowLoading />
  }
  return <div>グラフページ</div>
}

export default IsuGraph
