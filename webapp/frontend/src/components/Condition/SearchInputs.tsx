import { useState } from 'react'
import Button from '/@/components/UI/Button'
import Input from '/@/components/UI/Input'
import TimeInputs from './TimeInputs'

interface Props {
  query: string
  times: string[]
  search: (payload: { times: string[]; query: string }) => Promise<void>
}

const SearchInputs = ({ query, times, search }: Props) => {
  const [tmpQuery, setTmpQuery] = useState(query)
  const [tmpTimes, setTmpTimes] = useState(times)

  return (
    <div className="flex flex-wrap gap-6 items-end">
      <Input
        label="検索条件"
        value={tmpQuery}
        setValue={setTmpQuery}
        classname="flex-1"
      />
      <TimeInputs times={tmpTimes} setTimes={setTmpTimes} />
      <Button
        label="検索"
        onClick={() => search({ times: tmpTimes, query: tmpQuery })}
        disabled={!tmpQuery}
      />
    </div>
  )
}

export default SearchInputs
