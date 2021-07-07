import { Graph } from '../../lib/apis'
import Chart from 'react-apexcharts'
import { ApexOptions } from 'apexcharts'
import { useState, useEffect } from 'react'

interface Props {
  score: number
}

const Score = ({ score }: Props) => {
  const option: ApexOptions = {
    chart: {
      type: 'radialBar',
      offsetY: -20
    },
    plotOptions: {
      radialBar: {
        startAngle: -90,
        endAngle: 90,
        track: {
          background: '#e7e7e7',
          margin: 5
        },
        dataLabels: {
          name: {
            show: false
          },
          value: {
            offsetY: -2,
            fontSize: '22px'
          }
        }
      }
    },
    colors: ['#3dd47f'],
    series: [score]
  }

  return (
    <div>
      <Chart type="radialBar" options={option} series={option.series}></Chart>
    </div>
  )
}

export default Score
