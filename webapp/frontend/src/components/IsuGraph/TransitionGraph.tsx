import { Graph } from '../../lib/apis'
import Chart from 'react-apexcharts'
import { ApexOptions } from 'apexcharts'

interface Props {
  isuGraphs: Graph[]
}

const TransitionGraph = ({ isuGraphs }: Props) => {
  const data: number[] = []
  const categories: string[] = []
  isuGraphs.forEach(graph => {
    data.push(graph.data ? graph.data.score : 0)
    categories.push(graph.start_at)
  })

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
