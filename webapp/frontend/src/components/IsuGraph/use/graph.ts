import { useEffect, useState } from 'react'
import { GraphRequest, Graph } from '../../../lib/apis'

interface UseGraphResult {
  graphs: Graph[]
  transitionData: number[]
  sittingData: number[]
  timeCategories: string[]
  score: number
  day: string
}

const useGraph = (getGraphs: (req: GraphRequest) => Promise<Graph[]>) => {
  const [result, updateResult] = useState<UseGraphResult>({
    graphs: [],
    transitionData: [],
    sittingData: [],
    timeCategories: [],
    score: 0,
    day: ''
  })

  useEffect(() => {
    const fetchGraphs = async () => {
      const date = new Date()
      const graphs = await getGraphs({
        date: Date.parse(date.toLocaleDateString('ja-JP')) / 1000
      })
      const graphData = genGraphData(graphs)
      updateResult(state => ({
        ...state,
        graphs,
        transitionData: graphData.transitionData,
        sittingData: graphData.sittingData,
        timeCategories: graphData.timeCategories,
        score: graphData.score,
        day: date.toLocaleDateString('ja-JP')
      }))
    }
    fetchGraphs()
  }, [getGraphs, updateResult])

  const fetchGraphs = async (payload: { day: string }) => {
    // バリデーション
    const graphs = await getGraphs({ date: Date.parse(payload.day) / 1000 })
    const graphData = genGraphData(graphs)

    updateResult(state => ({
      ...state,
      loading: false,
      graphs,
      transitionData: graphData.transitionData,
      sittingData: graphData.sittingData,
      timeCategories: graphData.timeCategories,
      score: graphData.score,
      day: payload.day
    }))
  }

  return { ...result, fetchGraphs }
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
