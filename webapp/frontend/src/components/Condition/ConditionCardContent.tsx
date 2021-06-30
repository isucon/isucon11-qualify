import Conditions from './Conditions'
import SearchInputs from './SearchInputs'
import useConditions from './use/conditions'
import usePaging from './use/paging'

const ConditionCardContent = () => {
  const { conditions, setConditions } = useConditions()
  const { query, setQuery, times, setTimes, next, prev, page } = usePaging(
    conditions,
    setConditions
  )

  return (
    <div className="flex flex-col gap-2">
      <h2 className="text-xl font-bold">Condition</h2>
      <SearchInputs
        query={query}
        setQuery={setQuery}
        times={times}
        setTimes={setTimes}
      />
      <Conditions conditions={conditions} page={page} next={next} prev={prev} />
    </div>
  )
}

export default ConditionCardContent
