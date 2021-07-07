import { useEffect, useState } from 'react'
import { GraphRequest, Graph } from '../../../lib/apis'

const useGraph = (getGraphs: (req: GraphRequest) => Promise<Graph[]>) => {
  const [graphs, setGraphs] = useState<Graph[]>([])
  const [transitionData, setTransitionData] = useState<number[]>([])
  const [sittingData, setSittingData] = useState<number[]>([])
  const [timeCategories, setTimeCategories] = useState<string[]>([])
  const [score, setScore] = useState<number>(0)
  const [day, setDay] = useState<string>('')
  useEffect(() => {
    const fetchGraphs = async () => {
      const date = new Date()
      setGraphs(
        await getGraphs({
          date: Date.parse(date.toLocaleDateString('ja-JP')) / 1000
        })
      )

      const graphData = genGraphData(graphs)
      setTransitionData(graphData.transitionData)
      setSittingData(graphData.sittingData)
      setTimeCategories(graphData.timeCategories)
      setScore(graphData.score)
      setDay(date.toLocaleDateString('ja-JP'))
    }
    fetchGraphs()
  }, [
    getGraphs,
    setGraphs,
    graphs,
    setTransitionData,
    setSittingData,
    setTimeCategories,
    setScore,
    setDay
  ])

  const fetchGraphs = async (payload: { day: string }) => {
    // バリデーション
    setDay(payload.day)
    setGraphs(await getGraphs({ date: Date.parse(payload.day) / 1000 }))
    const graphData = genGraphData(graphs)
    setTransitionData(graphData.transitionData)
    setSittingData(graphData.sittingData)
    setTimeCategories(graphData.timeCategories)
    setScore(graphData.score)
  }

  return {
    graphs,
    transitionData,
    sittingData,
    timeCategories,
    score,
    day,
    fetchGraphs
  }
}

const dateToUnixtime = (date: Date) => {
  return Math.floor(date.getTime() / 1000)
}

const genGraphData = (graphs: Graph[]) => {
  const transitionData: number[] = []
  const sittingData: number[] = []
  const timeCategories: string[] = []
  let score = 0

  graphs.forEach(graph => {
    if (graph.data) {
      transitionData.push(graph.data.score)
      sittingData.push(graph.data.sitting)
      score += graph.data.score
    } else {
      transitionData.push(0)
      sittingData.push(0)
    }

    const date = new Date(graph.start_at * 1000)
    timeCategories.push(date.toLocaleTimeString('ja-JP'))
  })

  score /= graphs.length

  return {
    transitionData,
    sittingData,
    timeCategories,
    score
  }
}

export default useGraph
