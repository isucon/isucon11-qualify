import { Graph } from '../../lib/apis'
import Chart from 'react-apexcharts'
import { ApexOptions } from 'apexcharts'
import { useEffect, useState } from 'react'

interface Props {
  isuGraphs: Graph[]
}

const SittingGraph = ({ isuGraphs }: Props) => {
  const [data, setData] = useState<number[]>([])
  const [categories, setCategories] = useState<string[]>([])
  useEffect(() => {
    const load = () => {
      const tmpData: number[] = []
      const tmpCategories: string[] = []
      isuGraphs.forEach(graph => {
        tmpData.push(graph.data ? graph.data.sitting : 0)
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
      height: 100
    },
    colors: ['#ff6433'],
    series: [
      {
        type: 'heatmap',
        data: data
      }
    ],
    xaxis: {
      categories: categories
    },
    plotOptions: {
      heatmap: {
        colorScale: {
          ranges: [
            {
              from: 0,
              to: 20,
              color: '#d1d1d1'
            }
          ]
        }
      }
    }
  }

  return (
    <div>
      <Chart type="heatmap" options={option} series={option.series}></Chart>
    </div>
  )
}

export default SittingGraph
