import { useState } from 'react'
import Button from '../UI/Button'
import Input from '../UI/Input'

const SearchInputs = () => {
  const [query, setQuery] = useState('')
  const [time, setTime] = useState('')

  return (
    <div className="flex flex-wrap gap-6 items-center">
      <Input
        label="検索条件"
        value={query}
        setValue={setQuery}
        classname="flex-1"
      />
      <Input label="時間指定" value={time} setValue={setTime} />
      <Button label="検索" />
    </div>
  )
}

export default SearchInputs
