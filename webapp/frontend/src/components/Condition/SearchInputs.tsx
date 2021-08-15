import { useState } from 'react'
import ButtonSub from '/@/components/UI/ButtonSub'
import Input from '/@/components/UI/Input'
import TimeInputs from './TimeInputs'
import { ConditionRequest } from '/@/lib/apis'
import { getNowDate } from '/@/lib/date'

interface Props {
  query: ConditionRequest
  search: (params: ConditionRequest) => Promise<void>
}

const SearchInputs = ({ query, search }: Props) => {
  const [tmpConditionLevel, setTmpConditionLevel] = useState(
    query.condition_level
  )
  const [tmpStartTime, setTmpStartTime] = useState('')
  const [tmpEndTime, setTmpEndTime] = useState('')

  return (
    // string→Dateのパースはここでやる
    <div className="flex flex-wrap gap-6 items-end">
      <Input
        label="検索条件"
        value={tmpConditionLevel}
        setValue={setTmpConditionLevel}
        classname="flex-1"
      />
      <TimeInputs
        start_time={tmpStartTime}
        end_time={tmpEndTime}
        setStartTime={setTmpStartTime}
        setEndTime={setTmpEndTime}
      />
      <ButtonSub
        label="検索"
        onClick={() => {
          const start_time = new Date(tmpStartTime)
          const end_time = new Date(tmpEndTime)
          search({
            start_time: !isNaN(start_time.getTime()) ? start_time : new Date(0),
            end_time: !isNaN(end_time.getTime()) ? end_time : getNowDate(),
            condition_level: tmpConditionLevel
          })
        }}
      />
    </div>
  )
}

export default SearchInputs
