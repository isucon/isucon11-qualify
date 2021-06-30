import { useEffect } from 'react'
import { useState } from 'react'
import Conditions from './Conditions'
import SearchInputs from './SearchInputs'
import apis, { Condition, DEFAULT_CONDITION_LIMIT } from '../../lib/apis'

const ConditionCardContent = () => {
  const [conditions, setConditions] = useState<Condition[]>([])
  useEffect(() => {
    const fetchCondtions = async () => {
      setConditions(
        await apis.getConditions({
          cursor_end_time: new Date(),
          // 初回fetch時は'z'をセットすることで全件表示させてる
          cursor_jia_isu_uuid: 'z',
          condition_level: 'critical,warning,info'
        })
      )
    }
    fetchCondtions()
  }, [setConditions])

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
