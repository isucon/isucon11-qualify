import { Graph } from '../../lib/apis'
import Chart from 'react-apexcharts'
import { ApexOptions } from 'apexcharts'

interface Props {
  isuGraphs: Graph[]
}

const SittingGraph = ({ isuGraphs }: Props) => {
  const data: number[] = []
  const categories: string[] = []
  isuGraphs.forEach(graph => {
    data.push(graph.data ? graph.data.sitting : 0)
    categories.push(graph.start_at)
  })

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
