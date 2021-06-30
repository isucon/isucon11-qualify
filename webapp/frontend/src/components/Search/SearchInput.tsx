import { useState } from 'react'
import Button from '../UI/Button'
import Input from '../UI/Input'

interface Props {
  query: string
  search: (query: string) => Promise<void>
}

const SearchInput = ({ query, search }: Props) => {
  const [tmpQuery, setTmpQuery] = useState(query)

  return (
    <div className="flex gap-8 items-center mt-4 w-full">
      <Input
        label="検索条件"
        value={tmpQuery}
        setValue={setTmpQuery}
        classname="flex-1"
      />
      <Button label="検索" onClick={() => search(tmpQuery)} />
    </div>
  )
}

export default SearchInput
