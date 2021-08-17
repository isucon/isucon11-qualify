import { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import { useHistory, useLocation } from 'react-router-dom'
import { GraphRequest, Graph } from '/@/lib/apis'
import { dateToTimestamp, getNextDate, getPrevDate } from '/@/lib/date'

export interface IsuCondition {
  score: number
  count: number
  is_dirty: string
  is_overweight: string
  is_broken: string
}

interface GraphData {
  transitionData: number[]
  sittingData: number[]
  timeCategories: string[]
  day: string
  conditions: IsuCondition[]
}

const useGraph = (
  getGraphs: (req: GraphRequest) => Promise<Graph[]>,
  initialDate: Date
) => {
  const [result, updateResult] = useState<GraphData>({
    transitionData: [],
    sittingData: [],
    timeCategories: [],
    day: '',
    conditions: []
  })
  const [date, updateDate] = useState<Date>(initialDate)
  const history = useHistory()
  const location = useLocation()

  useEffect(() => {
    const fetchGraphs = async () => {
      const graphs = await getGraphs({ date: date })
      const graphData = genGraphData(graphs)
      updateResult(state => ({
        ...state,
        transitionData: graphData.transitionData,
        sittingData: graphData.sittingData,
        timeCategories: graphData.timeCategories,
        day: date.toLocaleDateString(),
        conditions: graphData.tooltipData
      }))
    }
    fetchGraphs()
  }, [getGraphs, updateResult, date, history, location.pathname])

  const specify = async (day: string) => {
    const date = new Date(day)
    if (isNaN(date.getTime())) {
      toast.error('日時の指定が不正です')
      return
    }
    replaceHistory()
    updateDate(date)
  }

  const prev = async () => {
    replaceHistory()
    updateDate(getPrevDate(date))
  }

  const next = async () => {
    replaceHistory()
    updateDate(getNextDate(date))
  }

  const replaceHistory = () =>
    history.replace(location.pathname + '?datetime=' + dateToTimestamp(date))

  return { ...result, specify, prev, next }
}

const genGraphData = (graphs: Graph[]) => {
  const transitionData: number[] = []
  const sittingData: number[] = []
  const timeCategories: string[] = []
  const tooltipData: IsuCondition[] = []

  graphs.forEach(graph => {
    if (graph.data) {
      transitionData.push(graph.data.score)
      sittingData.push(graph.data.percentage.sitting)
      tooltipData.push({
        score: graph.data.score,
        count: graph.condition_timestamps.length,
        is_dirty: `${graph.data.percentage.is_dirty}%`,
        is_overweight: `${graph.data.percentage.is_overweight}%`,
        is_broken: `${graph.data.percentage.is_broken}%`
      })
    } else {
      transitionData.push(0)
      sittingData.push(0)
      tooltipData.push({
        score: 0,
        count: 0,
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
