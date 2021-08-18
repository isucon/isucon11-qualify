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
        data: sittingData,
        name: ''
      }
    ],
    xaxis: {
      categories: timeCategories,
      labels: { show: false },
      axisBorder: { show: false },
      axisTicks: { show: false }
    },
    plotOptions: {
      heatmap: {
        colorScale: {
          ranges: [
            {
              from: 0,
              to: 20,
              color: '#ffe6df'
            },
            {
              from: 20,
              to: 40,
              color: '#ffb199'
            }
          ]
        }
      }
    }
  }

  return (
    <div
      style={{ transform: 'translateX(13px) translateY(-32px) scaleX(1.04)' }}
    >
      <Chart type="heatmap" options={option} series={option.series}></Chart>
    </div>
  )
}

export default SittingGraph
