import { Graph } from '../../lib/apis'
import Chart from 'react-apexcharts'
import { ApexOptions } from 'apexcharts'
import { useState } from 'react'
import { useEffect } from 'react'

interface Props {
  isuGraphs: Graph[]
}

const TransitionGraph = ({ isuGraphs }: Props) => {
  const [data, setData] = useState<number[]>([])
  const [categories, setCategories] = useState<string[]>([])

  useEffect(() => {
    const load = () => {
      const tmpData: number[] = []
      const tmpCategories: string[] = []
      isuGraphs.forEach(graph => {
        tmpData.push(graph.data ? graph.data.score : 0)
        const date = new Date(graph.start_at * 1000)
        tmpCategories.push(date.toLocaleTimeString('ja-JP'))
      })

      setData(tmpData)
      setCategories(tmpCategories)
    }
    load()
  }, [isuGraphs])

  const option: ApexOptions = {
    chart: {
      height: 350
    },
    colors: ['#008FFB'],
    series: [
      {
        type: 'line',
        data: data
      }
    ],
    xaxis: {
      categories: categories
    }
  }

  return (
    <div>
      <Chart type="line" options={option} series={option.series}></Chart>
    </div>
  )
}

export default TransitionGraph
