import { Dispatch, SetStateAction, useState } from 'react'
import { IoIosArrowBack, IoIosArrowForward } from 'react-icons/io'
import apis, { Condition } from '../../lib/apis'
import IconButton from '../UI/IconButton'
import ConditionDetail from './ConditionDetail'

interface Props {
  conditions: Condition[]
  setConditions: Dispatch<SetStateAction<Condition[]>>
}

const DEFAULT_CONDITION_LIMIT = 20

const Conditions = ({ conditions, setConditions }: Props) => {
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
  const isNextExist = conditions.length === DEFAULT_CONDITION_LIMIT
  const isPrevExist = page > 1

  return (
    <div className="flex flex-col gap-4 items-center">
      <div className="w-full border border-b-0 border-outline">
        {conditions.map((condition, i) => (
          <div className="border-b border-outline" key={i}>
            <ConditionDetail condition={condition} />
          </div>
        ))}
      </div>
      <div className="center flex gap-8">
        <IconButton disabled={!isPrevExist} onClick={prev}>
          <IoIosArrowBack size={24} />
        </IconButton>
        <div className="align-middle text-xl">{page}</div>
        <IconButton disabled={!isNextExist} onClick={next}>
          <IoIosArrowForward size={24} />
        </IconButton>
      </div>
    </div>
  )
}

export default Conditions
