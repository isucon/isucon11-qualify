import Chart from 'react-apexcharts'
import { ApexOptions } from 'apexcharts'

interface Props {
  sittingData: number[]
  timeCategories: string[]
}

const SittingGraph = ({ sittingData, timeCategories }: Props) => {
  const option: ApexOptions = {
    chart: {
      toolbar: {
        show: false
      }
    },
    colors: ['#ff6433'],
    series: [
      {
        type: 'heatmap',
        data: sittingData,
        name: ''
      }
    ],
    xaxis: {
      categories: timeCategories
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
      <div className="mb-3 text-xl">座った時間</div>
      <Chart type="heatmap" options={option} series={option.series}></Chart>
    </div>
  )
}

export default SittingGraph
