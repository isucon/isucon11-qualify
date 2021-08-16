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
    // string→Dateのバリデーション・パースはここでやる
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
          if (
            !tmpConditionLevel
              .split(',')
              .every(condition =>
                ['critical', 'warning', 'info'].includes(condition)
              )
          ) {
            alert(
              '検索条件には critical,warning,info のいずれか一つ以上をカンマ区切りで入力してください'
            )
          }
          const start_time = new Date(tmpStartTime)
          if (tmpStartTime && isNaN(start_time.getTime())) {
            alert('時間指定（since〜）が不正です')
            return
          }
          const end_time = new Date(tmpEndTime)
          if (tmpEndTime && isNaN(end_time.getTime())) {
            alert('時間指定（〜until）が不正です')
            return
          }
          search({
            condition_level: tmpConditionLevel,
            start_time: tmpStartTime ? start_time : undefined,
            end_time: tmpEndTime ? end_time : getNowDate()
          })
        }}
      />
    </div>
  )
}

export default SearchInputs
