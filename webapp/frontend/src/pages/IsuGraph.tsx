import { useCallback, useState } from 'react'
import apis, { Isu, GraphRequest } from '/@/lib/apis'
import TransitionGraph from '/@/components/IsuGraph/TransitionGraph'
import SittingGraph from '/@/components/IsuGraph/SittingGraph'
import useGraph from '/@/components/IsuGraph/use/graph'
import GraphNavigator from '/@/components/IsuGraph/GraphNavigator'
import NowLoading from '/@/components/UI/NowLoading'
import { useLocation } from 'react-router-dom'
import { getNowDate, timestampToDate } from '/@/lib/date'
import Legend from '/@/components/IsuGraph/Legend'

interface Props {
  isu: Isu
}

const IsuGraph = ({ isu }: Props) => {
  const [isLoading, setIsLoading] = useState(true)
  const getGraphs = useCallback(
    async (params: GraphRequest) => {
      setIsLoading(true)
      const res = await apis.getIsuGraphs(isu.jia_isu_uuid, params)
      setIsLoading(false)
      return res
    },
    [isu.jia_isu_uuid]
  )

  const queryParams = useLocation()
    .search.substring(1)
    .split('&')
    .reduce((acc, cur) => {
      acc[cur.split('=')[0]] = cur.split('=')[1]
      return acc
    }, {} as { [key: string]: string })
  let initialDate = getNowDate()
  initialDate = new Date(
    `${initialDate.getFullYear()}/${
      initialDate.getMonth() + 1
    }/${initialDate.getDate()}`
  )
  const initialDateTimestamp = Number(queryParams.datetime)
  if (!isNaN(initialDateTimestamp) && initialDateTimestamp > 0) {
    initialDate = timestampToDate(initialDateTimestamp)
  }

  const {
    transitionData,
    sittingData,
    timeCategories,
    day,
    conditions,
    specify,
    prev,
    next
  } = useGraph(getGraphs, initialDate)

  return (
    <div className="flex flex-col gap-12">
      <div className="flex justify-center w-full">
        <GraphNavigator prev={prev} next={next} specify={specify} day={day} />
      </div>
      <div className="relative flex flex-col gap-8">
        <div className="z-10">
          <TransitionGraph
            transitionData={transitionData}
            timeCategories={timeCategories}
            tooltipData={conditions}
            day={day}
          />
        </div>
        <div className="absolute top-0 w-full">
          <SittingGraph
            sittingData={sittingData}
            timeCategories={timeCategories}
          />
        </div>
        <div className="flex justify-center">
          <Legend />
        </div>
        {isLoading ? <NowLoading /> : null}
      </div>
    </div>
  )
}

export default IsuGraph
