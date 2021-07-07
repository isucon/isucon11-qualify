import apis, { Graph, Isu, GraphRequest } from '../../lib/apis'
import { useEffect, useCallback } from 'react'
import { useState } from 'react'
import NowLoading from '../UI/NowLoading'
import TransitionGraph from './TransitionGraph'
import SittingGraph from './SittingGraph'
import Score from './Score'
import DateInput from './DateInput'
import useGraph from './use/graph'

interface Props {
  isu: Isu
}

const IsuGraphCardContent = ({ isu }: Props) => {
  const getGraphs = useCallback(
    (params: GraphRequest) => {
      return apis.getIsuGraphs(isu.jia_isu_uuid, params)
    },
    [isu.jia_isu_uuid]
  )

  const {
    graphs,
    transitionData,
    sittingData,
    timeCategories,
    score,
    day,
    fetchGraphs
  } = useGraph(getGraphs)

  if (graphs.length === 0) {
    return <NowLoading />
  }
  return (
    <div>
      <DateInput day={day} fetchGraphs={fetchGraphs} />
      <TransitionGraph
        transitionData={transitionData}
        timeCategories={timeCategories}
      />
      <SittingGraph sittingData={sittingData} timeCategories={timeCategories} />
      <Score score={score} />
    </div>
  )
}

export default IsuGraphCardContent
