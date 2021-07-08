import { useEffect, useState } from 'react'
import { GraphRequest, Graph } from '../../../lib/apis'

export interface Tooltip {
  score: string
  is_dirty: string
  is_overweight: string
  is_broken: string
  missing_data: string
}

interface UseGraphResult {
  graphs: Graph[]
  transitionData: number[]
  sittingData: number[]
  timeCategories: string[]
  score: number
  day: string
  tooltipData: Tooltip[]
}

const useGraph = (getGraphs: (req: GraphRequest) => Promise<Graph[]>) => {
  const [result, updateResult] = useState<UseGraphResult>({
    graphs: [],
    transitionData: [],
    sittingData: [],
    timeCategories: [],
    score: 0,
    day: '',
    tooltipData: []
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
        day: date.toLocaleDateString('ja-JP'),
        tooltipData: graphData.tooltipData
      }))
    }
    fetchGraphs()
  }, [getGraphs, updateResult])

  const fetchGraphs = async (payload: { day: string }) => {
    const miliUnixtime = Date.parse(payload.day)
    if (isNaN(miliUnixtime)) {
      alert('日時の指定が不正です')
      return
    }

    const graphs = await getGraphs({ date: miliUnixtime / 1000 })
    const graphData = genGraphData(graphs)

    updateResult(state => ({
      ...state,
      loading: false,
      graphs,
      transitionData: graphData.transitionData,
      sittingData: graphData.sittingData,
      timeCategories: graphData.timeCategories,
      score: graphData.score,
      day: payload.day,
      tooltipData: graphData.tooltipData
    }))
  }

  return { ...result, fetchGraphs }
}

const genGraphData = (graphs: Graph[]) => {
  const transitionData: number[] = []
  const sittingData: number[] = []
  const timeCategories: string[] = []
  let score = 0
  const tooltipData: Tooltip[] = []

  graphs.forEach(graph => {
    if (graph.data) {
      transitionData.push(graph.data.score)
      sittingData.push(graph.data.sitting)
      score += graph.data.score
      tooltipData.push({
        score: graph.data.score.toString(),
        is_dirty: graph.data.detail['is_dirty']
          ? graph.data.detail['is_dirty'].toString()
          : '-',
        is_overweight: graph.data.detail['is_overweight']
          ? graph.data.detail['is_overweight'].toString()
          : '-',
        is_broken: graph.data.detail['is_broken']
          ? graph.data.detail['is_broken'].toString()
          : '-',
        missing_data: graph.data.detail['missing_data']
          ? graph.data.detail['missing_data'].toString()
          : '-'
      })
    } else {
      transitionData.push(0)
      sittingData.push(0)
      tooltipData.push({
        score: '-',
        is_dirty: '-',
        is_overweight: '-',
        is_broken: '-',
        missing_data: '-'
      })
    }

    const date = new Date(graph.start_at * 1000)
    timeCategories.push(
      date.toLocaleTimeString('ja-JP', { hour: '2-digit', minute: '2-digit' })
    )
  })

  score = Math.floor(score / graphs.length)

  return {
    transitionData,
    sittingData,
    timeCategories,
    score,
    tooltipData
  }
}

export default useGraph
