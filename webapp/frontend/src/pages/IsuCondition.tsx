import { useCallback, useState } from 'react'
import apis, { ConditionRequest, Isu } from '/@/lib/apis'
import ConditionNavigator from '/@/components/Condition/ConditionNavigator'
import ConditionList from '/@/components/Condition/ConditionList'
import SearchInputs from '/@/components/Condition/SearchInputs'
import usePagingCondition from '/@/components/Condition/use/paging'
import NowLoading from '/@/components/UI/NowLoading'
import { useLocation } from 'react-router-dom'
import { getNowDate, timestampToDate } from '/@/lib/date'

interface Props {
  isu: Isu
}

const IsuCondition = ({ isu }: Props) => {
  const [isLoading, setIsLoading] = useState(true)
  const getConditions = useCallback(
    async (params: ConditionRequest) => {
      setIsLoading(true)
      const res = await apis.getIsuConditions(isu.jia_isu_uuid, params)
      setIsLoading(false)
      return res
    },
    [isu]
  )

  const queryParams = useLocation()
    .search.substring(1)
    .split('&')
    .reduce((acc, cur) => {
      acc[cur.split('=')[0]] = cur.split('=')[1]
      return acc
    }, {} as { [key: string]: string })
  const condition_level = queryParams.condition_level ?? 'critical,warning,info'
  let start_time = undefined
  let end_time = getNowDate()
  const start_timestamp = Number(queryParams.start_time)
  if (!isNaN(start_timestamp) && start_timestamp > 0) {
    start_time = timestampToDate(start_timestamp)
  }
  const end_timestamp = Number(queryParams.end_time)
  if (!isNaN(end_timestamp) && end_timestamp > (start_timestamp || 0)) {
    end_time = timestampToDate(end_timestamp)
  }

  const { conditions, query, search, next, prev, page } = usePagingCondition(
    getConditions,
    { condition_level, start_time, end_time }
  )

  return (
    <div className="flex flex-col gap-8">
      <SearchInputs query={query} search={search} />
      <div className="relative flex flex-col gap-4 items-center">
        <ConditionList conditions={conditions} />
        {isLoading ? <NowLoading top /> : null}
        <ConditionNavigator
          conditions={conditions}
          page={page}
          next={next}
          prev={prev}
        />
      </div>
    </div>
  )
}

export default IsuCondition
