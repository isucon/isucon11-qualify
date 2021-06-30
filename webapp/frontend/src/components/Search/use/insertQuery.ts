import { Dispatch, RefObject, SetStateAction } from 'react'

const useInsertQuery = (
  inputRef: RefObject<HTMLInputElement>,
  query: string,
  setQuery: Dispatch<SetStateAction<string>>
) => {
  const insert = (key: string) => {
    if (inputRef.current) {
      const newQuery = `${query} ${key}:""`
      inputRef.current.value = newQuery
      inputRef.current.focus()
      inputRef.current.setSelectionRange(
        newQuery.length - 1,
        newQuery.length - 1
      )
      setQuery(newQuery)
    } else {
      throw 'inputRef.current is falthy'
    }
  }
  return { insert }
}

export default useInsertQuery
