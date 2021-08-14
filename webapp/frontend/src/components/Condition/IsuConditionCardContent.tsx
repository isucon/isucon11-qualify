import { useCallback } from 'react'
import apis, { ConditionRequest, Isu } from '/@/lib/apis'
import ConditionNavigator from './ConditionNavigator'
import Conditions from './Conditions'
import SearchInputs from './SearchInputs'
import usePaging from './use/paging'
import { useState } from 'react'
import NowLoadingOverlay from '/@/components/UI/NowLoadingOverlay'

interface Props {
  isu: Isu
}
const IsuConditionCardContent = ({ isu }: Props) => {
  const [isLoading, setIsLoading] = useState(false)
  const getConditions = useCallback(
    async (params: ConditionRequest) => {
      setIsLoading(true)
      const res = await apis.getIsuConditions(isu.jia_isu_uuid, params)
      setIsLoading(false)
      return res
    },
    [isu]
  )
  const { conditions, query, times, search, next, prev, page } =
    usePaging(getConditions)

  return (
    <div className="flex flex-col gap-2">
      <SearchInputs query={query} times={times} search={search} />
      <div className="flex flex-col gap-4 items-center">
        <div className="relative w-full">
          <Conditions conditions={conditions} />
          {isLoading ? <NowLoadingOverlay /> : null}
        </div>
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

export default IsuConditionCardContent
