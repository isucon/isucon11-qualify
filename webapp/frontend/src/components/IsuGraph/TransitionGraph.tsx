import Chart from 'react-apexcharts'
import { ApexOptions } from 'apexcharts'
import { IsuCondition } from './use/graph'

interface Props {
  transitionData: number[]
  timeCategories: string[]
  tooltipData: IsuCondition[]
}

const TransitionGraph = ({
  transitionData,
  timeCategories,
  tooltipData: tooltopData
}: Props) => {
  const option: ApexOptions = {
    chart: {
      toolbar: {
        show: false
      }
    },
    grid: {
      yaxis: {
        lines: { show: false }
      }
    },
    colors: ['#008FFB'],
    series: [
      {
        type: 'line',
        data: transitionData
      }
    ],
    xaxis: {
      categories: timeCategories,
      offsetY: 8
    },
    yaxis: {
      min: 0,
      labels: {
        offsetX: -16
      }
    },
    tooltip: {
      custom: ({ dataPointIndex }) => {
        return genTooltipCard(tooltopData[dataPointIndex])
      }
    }
  }

  return (
    <div>
      <Chart type="line" options={option} series={option.series}></Chart>
    </div>
  )
}

const genTooltipCard = (tooltip: IsuCondition) => {
  return `
  <div class="flex flex-col px-3 py-1 text-primary">
    <div class="flex flex-row">
      <div class="w-25">score</div>
      <div>${tooltip.score}</div>
    </div>
    <div class="flex flex-row">
      <div class="w-25">is_dirty</div>
      <div>${tooltip.is_dirty}</div>
    </div>
    <div class="flex flex-row">
      <div class="w-25">is_overweight</div>
      <div>${tooltip.is_overweight}</div>
    </div>
    <div class="flex flex-row">
      <div class="w-25">is_broken</div>
      <div>${tooltip.is_broken}</div>
    </div>
  </div>`
}

export default TransitionGraph
