import { useEffect, useState } from 'react'
import { useHistory, useLocation } from 'react-router-dom'
import {
  Condition,
  ConditionRequest,
  DEFAULT_CONDITION_LIMIT
} from '/@/lib/apis'
import { dateToTimestamp } from '/@/lib/date'

const usePagingCondition = (
  getConditions: (req: ConditionRequest) => Promise<Condition[]>,
  initialQuery: ConditionRequest
) => {
  const [query, setQuery] = useState<ConditionRequest>(initialQuery)
  const [page, setPage] = useState(1)
  const [conditions, setConditions] = useState<Condition[]>([])
  // 1-indexed
  const [cache, setCache] = useState<
    { query: ConditionRequest; conditions: Condition[] }[]
  >([])

  useEffect(() => {
    const fetchCondtions = async () => {
      const conditions = await getConditions(query)
      setConditions(conditions)
    }
    fetchCondtions()
  }, [getConditions, setConditions, query])

  const search = async (params: ConditionRequest) => {
    replaceHistory(params)
    setQuery(params)
    setPage(1)
    setCache([
      { query: params, conditions: [] }, // 埋める用
      { query: params, conditions }
    ])
  }

  const next = async () => {
    const newQuery = getNextRequestParams()
    replaceHistory(newQuery)
    setQuery(newQuery)
    setPage(page + 1)
    if (!cache[page]) {
      cache[page] = { query, conditions }
      setCache(cache)
    }
  }
  const prev = async () => {
    replaceHistory(cache[page - 1].query)
    setConditions(cache[page - 1].conditions)
    setPage(page - 1)
  }

  const getNextRequestParams = (): ConditionRequest => {
    const start_time = query.start_time
    const end_time = conditions[DEFAULT_CONDITION_LIMIT - 1].date
    return {
      condition_level: query.condition_level,
      start_time,
      end_time
    }
  }

  const history = useHistory()
  const location = useLocation()

  const replaceHistory = (query: ConditionRequest) => {
    const newLocation = `${location.pathname}?condition_level=${
      query.condition_level
    }${
      !isNaN(query.end_time.getTime())
        ? '&end_time=' + dateToTimestamp(query.end_time)
        : ''
    }${
      query.start_time && !isNaN(query.start_time.getTime())
        ? '&start_time=' + dateToTimestamp(query.start_time)
        : ''
    }`
    history.replace(newLocation)
  }

  return { conditions, query, search, page, next, prev }
}

export default usePagingCondition
