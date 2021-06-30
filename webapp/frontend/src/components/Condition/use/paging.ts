import { Dispatch, SetStateAction, useState } from 'react'
import apis, { Condition, DEFAULT_CONDITION_LIMIT } from '../../../lib/apis'

const usePaging = (
  conditions: Condition[],
  setConditions: Dispatch<SetStateAction<Condition[]>>
) => {
  const [query, setQuery] = useState('')
  const [times, setTimes] = useState(['', ''])
  const [cache, setCache] = useState<Condition[][]>([[]])
  const [page, setPage] = useState(1)
  const next = async () => {
    if (!cache[page]) {
      cache[page] = conditions
      setCache(cache)
    }
    setConditions(
      await apis.getConditions({
        cursor_end_time: new Date(
          conditions[DEFAULT_CONDITION_LIMIT - 1].timestamp
        ),
        cursor_jia_isu_uuid:
          conditions[DEFAULT_CONDITION_LIMIT - 1].jia_isu_uuid,
        condition_level: 'critical,warning,info'
      })
    )
    setPage(page + 1)
  }
  const prev = async () => {
    setConditions(cache[page - 1])
    setPage(page - 1)
  }
  const search = async (payload: { times: string[]; query: string }) => {
    setQuery(payload.query)
    setTimes(payload.times)
    // setConditions()
  }

  return { query, times, search, page, next, prev }
}

export default usePaging
