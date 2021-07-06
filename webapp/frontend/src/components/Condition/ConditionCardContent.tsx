import apis, { DEFAULT_CONDITION_LIMIT } from '../../lib/apis'
import PagingNavigator from '../UI/PagingNavigator'
import Conditions from './Conditions'
import SearchInputs from './SearchInputs'
import usePaging from './use/paging'

const ConditionCardContent = () => {
  const { conditions, query, times, search, next, prev, page } = usePaging(
    apis.getConditions
  )

  return (
    <div className="flex flex-col gap-2">
      <h2 className="text-xl font-bold">Condition</h2>
      <SearchInputs query={query} times={times} search={search} />
      <div className="flex flex-col gap-4 items-center">
        <Conditions conditions={conditions} />
        <PagingNavigator
          length={conditions.length}
          maxLength={DEFAULT_CONDITION_LIMIT}
          page={page}
          next={next}
          prev={prev}
        />
      </div>
    </div>
  )
}

export default ConditionCardContent
