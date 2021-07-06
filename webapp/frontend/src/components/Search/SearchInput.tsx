import { useRef } from 'react'
import { useState } from 'react'
import Button from '../UI/Button'
import Input from '../UI/Input'
import HelperModal from './HelperModal'
import useHelperModal from './use/helperModal'
import useInsertQuery from './use/insertQuery'

interface Props {
  query: string
  search: (query: string) => Promise<void>
}

const placeholder = `name:isuname character:Adamant catalog_name:isu1 min_limit_weight:20 max_limit_weight:50 catalog_tags:"tag1 tag2"`

const SearchInput = ({ query, search }: Props) => {
  const [tmpQuery, setTmpQuery] = useState(query)
  const inputRef = useRef<HTMLInputElement>(null)
  const { insert } = useInsertQuery(inputRef, tmpQuery, setTmpQuery)
  const { isOpen, toggle, rect } = useHelperModal(inputRef)

  return (
    <div className="flex gap-8 items-center mt-4 w-full">
      <Input
        label="検索条件"
        value={tmpQuery}
        setValue={setTmpQuery}
        classname="flex-1"
        inputProps={{ placeholder: placeholder, ref: inputRef }}
      />
      <Button label="検索" onClick={() => search(tmpQuery)} />
      <HelperModal
        isOpen={isOpen}
        toggle={toggle}
        insertOption={insert}
        rect={rect}
      />
    </div>
  )
}

export default SearchInput
