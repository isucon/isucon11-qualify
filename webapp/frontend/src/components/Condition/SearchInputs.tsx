import { useState } from 'react'
import Button from '/@/components/UI/Button'
import Input from '/@/components/UI/Input'
import TimeInputs from './TimeInputs'
import { ConditionRequest } from '/@/lib/apis'
import { getNowDate, getConditionTime } from '/@/lib/date'
import toast from 'react-hot-toast'

interface Props {
  query: ConditionRequest
  search: (params: ConditionRequest) => Promise<void>
}

const SearchInputs = ({ query, search }: Props) => {
  const initStartTime = query.start_time
    ? getConditionTime(query.start_time)
    : ''
  const initEndTime = query.end_time ? getConditionTime(query.end_time) : ''

  const [tmpConditionLevel, setTmpConditionLevel] = useState(
    query.condition_level
  )
  const [tmpStartTime, setTmpStartTime] = useState(initStartTime)
  const [tmpEndTime, setTmpEndTime] = useState(initEndTime)

  return (
    // string→Dateのバリデーション・パースはここでやる
    <div className="flex flex-wrap gap-6 items-end">
      <Input
        label="検索条件"
        value={tmpConditionLevel}
        setValue={setTmpConditionLevel}
        customClass="flex-1"
      />
      <TimeInputs
        start_time={tmpStartTime}
        end_time={tmpEndTime}
        setStartTime={setTmpStartTime}
        setEndTime={setTmpEndTime}
      />
      <Button
        customClass="px-3 py-1 h-8 leading-4 border border-primary rounded"
        label="検索"
        disabled={!tmpConditionLevel}
        onClick={() => {
          if (
            !tmpConditionLevel
              .split(',')
              .every(condition =>
                ['critical', 'warning', 'info'].includes(condition)
              )
          ) {
            toast.error(
              '検索条件には critical,warning,info のいずれか一つ以上をカンマ区切りで入力してください'
            )
            return
          }
          const start_time = new Date(tmpStartTime)
          if (tmpStartTime && isNaN(start_time.getTime())) {
            toast.error('時間指定（start_time〜）が不正です')
            return
          }
          const end_time = new Date(tmpEndTime)
          if (tmpEndTime && isNaN(end_time.getTime())) {
            toast.error('時間指定（〜end_time）が不正です')
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
