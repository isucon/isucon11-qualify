import { Graph } from '../../lib/apis'
import Chart from 'react-apexcharts'
import { ApexOptions } from 'apexcharts'
import { useState, useEffect } from 'react'

interface Props {
  isuGraphs: Graph[]
}

const Score = ({ isuGraphs }: Props) => {
  const [score, setScore] = useState<number>(0)
  useEffect(() => {
    const calcScore = () => {
      let tmpScore = 0
      isuGraphs.forEach(isuGraph => {
        tmpScore += isuGraph.data ? isuGraph.data.score : 0
      })
      tmpScore /= isuGraphs.length
      return tmpScore
    }
    setScore(calcScore())
  }, [isuGraphs])

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
