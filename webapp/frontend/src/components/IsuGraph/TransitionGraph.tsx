import Chart from 'react-apexcharts'
import { ApexOptions } from 'apexcharts'
import { IsuCondition } from './use/graph'
import colors from 'windicss/colors'
import { useHistory, useParams } from 'react-router-dom'
import { dateToTimestamp } from '/@/lib/date'

interface Props {
  transitionData: number[]
  timeCategories: string[]
  tooltipData: IsuCondition[]
  day: string
}

const TransitionGraph = ({
  transitionData,
  timeCategories,
  tooltipData: tooltopData,
  day
}: Props) => {
  const history = useHistory()
  const { id } = useParams<{ id: string }>()
  const option: ApexOptions = {
    chart: {
      toolbar: {
        show: false
      },
      zoom: {
        enabled: false
      },
      events: {
        click: (e, chart, config) => {
          if (!config && typeof config.dataPointIndex !== 'number') {
            return
          }
          const index = config.dataPointIndex
          if (transitionData.length <= index) {
            return
          }
          const date = new Date(day)
          if (isNaN(date.getTime())) {
            return
          }
          const startTime = dateToTimestamp(date)
          const endTime = startTime + 60 * 60 * index
          history.push(
            `/isu/${id}/condition?end_time=${endTime}&start_time=${startTime}`
          )
        }
      }
    },
    grid: {
      yaxis: {
        lines: { show: false }
      }
    },
    colors: [colors.blue[500]],
    series: [
      {
        type: 'line',
        data: transitionData
      }
    ],
    xaxis: {
      categories: timeCategories,
      offsetY: 8,
      labels: {
        rotateAlways: true
      }
    },
    yaxis: {
      min: 0,
      max: 100,
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
