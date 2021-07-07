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
      height: 350
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
    }
  }

  return (
    <div>
      <Chart type="line" options={option} series={option.series}></Chart>
    </div>
  )
}

export default TransitionGraph
