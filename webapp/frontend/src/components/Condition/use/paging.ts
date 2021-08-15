import { useEffect, useState } from 'react'
import { useHistory, useLocation } from 'react-router-dom'
import {
  Condition,
  ConditionRequest,
  DEFAULT_CONDITION_LIMIT
} from '/@/lib/apis'
import { dateToTimestamp, getNowDate } from '/@/lib/date'

const usePagingCondition = (
  getConditions: (req: ConditionRequest) => Promise<Condition[]>
) => {
  const [query, setQuery] = useState<ConditionRequest>({
    condition_level: 'critical,warning,info',
    end_time: getNowDate()
  })
  const [page, setPage] = useState(1)
  const [conditions, setConditions] = useState<Condition[]>([])
  // 1-indexed
  const [cache, setCache] = useState<Condition[][]>([[]])

  const history = useHistory()
  const location = useLocation()

  useEffect(() => {
    const fetchCondtions = async () => {
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
      history.push(newLocation)
      const conditions = await getConditions(query)
      setConditions(conditions)
    }
    fetchCondtions()
  }, [getConditions, setConditions, query, location.pathname, history])

  const search = async (params: ConditionRequest) => {
    setQuery(params)
    setPage(1)
    setCache([[], conditions])
  }

  const next = async () => {
    setQuery(getNextRequestParams())
    setPage(page + 1)
    if (!cache[page]) {
      cache[page] = conditions
      setCache(cache)
    }
  }
  const prev = async () => {
    setConditions(cache[page - 1])
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

  return { conditions, query, search, page, next, prev }
}

export default usePagingCondition
