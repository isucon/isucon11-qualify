import { useEffect, useState } from 'react'
import {
  Condition,
  ConditionRequest,
  DEFAULT_CONDITION_LIMIT
} from '../../../lib/apis'

const usePaging = (
  getConditions: (req: ConditionRequest) => Promise<Condition[]>
) => {
  const [conditions, setConditions] = useState<Condition[]>([])
  useEffect(() => {
    const fetchCondtions = async () => {
      setConditions(
        await getConditions({
          cursor_end_time: Math.floor(new Date().getTime() / 1000),
          // 初回fetch時は'z'をセットすることで全件表示させてる
          cursor_jia_isu_uuid: 'z',
          condition_level: 'critical,warning,info'
        })
      )
    }
    fetchCondtions()
  }, [getConditions, setConditions])

  const [query, setQuery] = useState('critical,warning,info')
  const [times, setTimes] = useState(['', ''])
  const [cache, setCache] = useState<Condition[][]>([[]])
  const [page, setPage] = useState(1)
  const next = async () => {
    if (!cache[page]) {
      cache[page] = conditions
      setCache(cache)
    }
    const params = getNextRequestParams(
      conditions[DEFAULT_CONDITION_LIMIT - 1].jia_isu_uuid
    )
    if (!params) {
      return
    }
    setConditions(await getConditions(params))
    setPage(page + 1)
  }
  const prev = async () => {
    setConditions(cache[page - 1])
    setPage(page - 1)
  }
  const search = async (payload: { times: string[]; query: string }) => {
    if (payload.query) {
      if (!validateQuery(payload.query)) {
        return
      }
    }

    let start_time: Date
    if (payload.times[0]) {
      const date = validateTime(payload.times[0] + 'Z')
      if (date) {
        start_time = date
      } else {
        alert('時間指定の下限が不正です')
        return
      }
    } else {
      start_time = new Date(0)
    }

    let cursor_end_time: Date
    if (payload.times[1]) {
      const date = validateTime(payload.times[1] + 'Z')
      if (date) {
        cursor_end_time = date
      } else {
        alert('時間指定の上限が不正です')
        return
      }
    } else {
      cursor_end_time = new Date()
    }

    setQuery(payload.query)
    setTimes(payload.times)

    const params = {
      start_time: start_time.getTime(),
      cursor_end_time: cursor_end_time.getTime(),
      condition_level: payload.query,
      cursor_jia_isu_uuid: 'z'
    }
    setConditions(await getConditions(params))
    setPage(1)
    setCache([[]])
  }

  // 初回fetch時はcursor_jia_isu_uuidに'z'をセットすることで全件表示させてる
  const getNextRequestParams = (cursor_jia_isu_uuid = 'z') => {
    const start_time = times[0] ? new Date(times[0] + 'Z') : new Date(0)
    const cursor_end_time = times[1]
      ? new Date(times[1] + 'Z')
      : new Date(conditions[DEFAULT_CONDITION_LIMIT - 1].timestamp)

    return {
      cursor_end_time: Math.floor(cursor_end_time.getTime() / 1000),
      start_time: Math.floor(start_time.getTime() / 1000),
      cursor_jia_isu_uuid,
      condition_level: query
    }
  }

  return { conditions, query, times, search, page, next, prev }
}

const validateQuery = (query: string) => {
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

const validateTime = (time: string) => {
  const date = new Date(time)
  if (isNaN(date.getTime())) {
    return false
  }
  return date
}

export default usePaging
