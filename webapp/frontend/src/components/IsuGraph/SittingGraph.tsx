import { Graph } from '../../lib/apis'
import Chart from 'react-apexcharts'
import { ApexOptions } from 'apexcharts'
import { useEffect, useState } from 'react'

interface Props {
  sittingData: number[]
  timeCategories: string[]
}

const SittingGraph = ({ sittingData, timeCategories }: Props) => {
  const option: ApexOptions = {
    chart: {
      height: 100
    },
    colors: ['#ff6433'],
    series: [
      {
        type: 'heatmap',
        data: sittingData
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
      <Chart type="heatmap" options={option} series={option.series}></Chart>
    </div>
  )
}

export default SittingGraph
