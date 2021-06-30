import apis from '../../lib/apis'
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
      <Conditions conditions={conditions} page={page} next={next} prev={prev} />
    </div>
  )
}

export default ConditionCardContent
