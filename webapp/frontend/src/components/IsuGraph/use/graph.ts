import { useEffect, useState } from 'react'
import { GraphRequest, Graph } from '../../../lib/apis'
import { getNextDate, getPrevDate, getTodayDate } from '../../../lib/date'

export interface Tooltip {
  score: string
  is_dirty: string
  is_overweight: string
  is_broken: string
}

interface UseGraphResult {
  graphs: Graph[]
  transitionData: number[]
  sittingData: number[]
  timeCategories: string[]
  day: string
  tooltipData: Tooltip[]
}

const useGraph = (getGraphs: (req: GraphRequest) => Promise<Graph[]>) => {
  const [result, updateResult] = useState<UseGraphResult>({
    graphs: [],
    transitionData: [],
    sittingData: [],
    timeCategories: [],
    day: '',
    tooltipData: []
  })
  const [date, updateDate] = useState<Date>(getTodayDate())

  useEffect(() => {
    const fetchGraphs = async () => {
      const graphs = await getGraphs({
        date: date
      })
      const graphData = genGraphData(graphs)
      updateResult(state => ({
        ...state,
        graphs,
        transitionData: graphData.transitionData,
        sittingData: graphData.sittingData,
        timeCategories: graphData.timeCategories,
        day: date.toLocaleDateString(),
        tooltipData: graphData.tooltipData
      }))
    }
    fetchGraphs()
  }, [getGraphs, updateResult, date])

  const innerFetchGraphs = async () => {
    const graphs = await getGraphs({ date: date })
    const graphData = genGraphData(graphs)

    updateResult(state => ({
      ...state,
      loading: false,
      graphs,
      transitionData: graphData.transitionData,
      sittingData: graphData.sittingData,
      timeCategories: graphData.timeCategories,
      day: date.toLocaleTimeString(),
      tooltipData: graphData.tooltipData
    }))
  }

  const fetchGraphs = async (payload: { day: string }) => {
    const date = new Date(payload.day)
    if (isNaN(date.getTime())) {
      alert('日時の指定が不正です')
      return
    }

    updateDate(date)
    innerFetchGraphs()
  }

  const prev = async () => {
    updateDate(getPrevDate(date))
    innerFetchGraphs()
  }

  const next = async () => {
    updateDate(getNextDate(date))
    innerFetchGraphs()
  }

  return { ...result, fetchGraphs, prev, next }
}

const genGraphData = (graphs: Graph[]) => {
  const transitionData: number[] = []
  const sittingData: number[] = []
  const timeCategories: string[] = []
  const tooltipData: Tooltip[] = []

  graphs.forEach(graph => {
    if (graph.data) {
      transitionData.push(graph.data.score)
      sittingData.push(graph.data.percentage.sitting)
      tooltipData.push({
        score: graph.data.score.toString(),
        is_dirty: `${graph.data.percentage.is_dirty}%`,
        is_overweight: `${graph.data.percentage.is_overweight}%`,
        is_broken: `${graph.data.percentage.is_broken}%`
      })
    } else {
      transitionData.push(0)
      sittingData.push(0)
      tooltipData.push({
        score: '-',
        is_dirty: '-',
        is_overweight: '-',
        is_broken: '-'
      })
    }

    timeCategories.push(
      graph.start_at.toLocaleTimeString('ja-JP', {
        hour: '2-digit',
        minute: '2-digit'
      })
    )
  })

  return {
    transitionData,
    sittingData,
    timeCategories,
    tooltipData
  }
}

export default useGraph
