import { Dispatch, SetStateAction } from 'react'
import Button from '../UI/Button'
import Input from '../UI/Input'
import TimeInputs from './TimeInputs'

interface Props {
  query: string
  setQuery: Dispatch<SetStateAction<string>>
  times: string[]
  setTimes: Dispatch<React.SetStateAction<string[]>>
}

const SearchInputs = ({ query, setQuery, times, setTimes }: Props) => {
  return (
    <div className="flex flex-wrap gap-6 items-center">
      <Input
        label="検索条件"
        value={query}
        setValue={setQuery}
        classname="flex-1"
      />
      <TimeInputs times={times} setTimes={setTimes} />
      <Button label="検索" />
    </div>
  )
}

export default SearchInputs
