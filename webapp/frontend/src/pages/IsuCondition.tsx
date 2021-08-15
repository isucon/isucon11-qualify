import { useCallback, useState } from 'react'
import apis, { ConditionRequest, Isu } from '/@/lib/apis'
import ConditionNavigator from '/@/components/Condition/ConditionNavigator'
import ConditionList from '/@/components/Condition/ConditionList'
import SearchInputs from '/@/components/Condition/SearchInputs'
import usePagingCondition from '/@/components/Condition/use/paging'
import NowLoading from '/@/components/UI/NowLoading'

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
  const { conditions, query, search, next, prev, page } =
    usePagingCondition(getConditions)

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
