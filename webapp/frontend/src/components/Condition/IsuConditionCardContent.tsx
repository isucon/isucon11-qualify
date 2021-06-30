import { useCallback } from 'react'
import apis, { ConditionRequest, Isu } from '../../lib/apis'
import Conditions from './Conditions'
import SearchInputs from './SearchInputs'
import usePaging from './use/paging'

interface Props {
  isu: Isu
}
const IsuConditionCardContent = ({ isu }: Props) => {
  const getConditions = useCallback(
    (params: ConditionRequest) => {
      return apis.getIsuConditions(isu.jia_isu_uuid, params)
    },
    [isu]
  )
  const { conditions, query, times, search, next, prev, page } =
    usePaging(getConditions)

  return (
    <div className="flex flex-col gap-2">
      <SearchInputs query={query} times={times} search={search} />
      <Conditions conditions={conditions} page={page} next={next} prev={prev} />
    </div>
  )
}

export default IsuConditionCardContent
