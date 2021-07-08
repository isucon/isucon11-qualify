import { Graph } from '../../lib/apis'
import Chart from 'react-apexcharts'
import { ApexOptions } from 'apexcharts'
import { useState } from 'react'
import { useEffect } from 'react'

interface Props {
  transitionData: number[]
  timeCategories: string[]
}

const TransitionGraph = ({ transitionData, timeCategories }: Props) => {
  const option: ApexOptions = {
    chart: {
      toolbar: {
        show: false
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
      categories: timeCategories
    },
    yaxis: {
      max: 100
    }
  }

  return (
    <div>
      <div className="mb-3 text-xl">推移</div>
      <Chart type="line" options={option} series={option.series}></Chart>
    </div>
  )
}

export default TransitionGraph
