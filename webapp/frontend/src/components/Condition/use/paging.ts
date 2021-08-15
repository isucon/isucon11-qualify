import { useEffect, useState } from 'react'
import {
  Condition,
  ConditionRequest,
  DEFAULT_CONDITION_LIMIT
} from '/@/lib/apis'
import { getNowDate } from '/@/lib/date'

const usePagingCondition = (
  getConditions: (req: ConditionRequest) => Promise<Condition[]>
) => {
  const [conditions, setConditions] = useState<Condition[]>([])

  const [query, setQuery] = useState<ConditionRequest>({
    condition_level: 'critical,warning,info',
    end_time: getNowDate()
  })

  // 1-indexed
  const [cache, setCache] = useState<Condition[][]>([[]])
  const [page, setPage] = useState(1)

  useEffect(() => {
    const fetchCondtions = async () => {
      const conditions = await getConditions(query)
      setConditions(conditions)
    }
    fetchCondtions()
  }, [getConditions, setConditions, query])

  const search = async (params: ConditionRequest) => {
    if (params.condition_level) {
      if (!validateConditionLevel(params.condition_level)) {
        return
      }
    }

    if (params.start_time) {
      if (!isNaN(params.start_time.getTime())) {
        setQuery(query => ({
          ...query,
          start_time: params.start_time
        }))
      } else {
        alert('時間指定の下限が不正です')
        return
      }
    }

    if (params.end_time) {
      if (!isNaN(params.end_time.getTime())) {
        setQuery(query => ({
          ...query,
          end_time: params.end_time
        }))
      } else {
        alert('時間指定の上限が不正です')
        return
      }
    }

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
    const start_time = query.start_time ? query.start_time : new Date(0)
    const end_time = conditions[DEFAULT_CONDITION_LIMIT - 1].date

    return {
      end_time,
      start_time,
      condition_level: query.condition_level
    }
  }

  return { conditions, query, search, page, next, prev }
}

const validateConditionLevel = (query: string) => {
  const splitQuery = query.split(',')
  for (const sq of splitQuery) {
    if (!['critical', 'warning', 'info'].includes(sq)) {
      alert(
        '検索条件には critical,warning,info のいずれか一つ以上をカンマ区切りで入力してください'
      )
      return false
    }
  }
  return true
}

export default usePagingCondition
