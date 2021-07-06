import { Dispatch, RefObject, SetStateAction, useCallback } from 'react'
import { parseQuery } from './parseQuery'

const useInsertQuery = (
  inputRef: RefObject<HTMLInputElement>,
  query: string,
  setQuery: Dispatch<SetStateAction<string>>
) => {
  const insert = (key: string) => {
    if (inputRef.current) {
      const params = parseQuery(query)
      if (params[key]) {
        setCursor(params[key].valueNextIndex)
      } else {
        const newQuery = `${query} ${key}:""`
        inputRef.current.value = newQuery
        setCursor(newQuery.length - 1)
        setQuery(newQuery)
      }
    } else {
      throw 'inputRef.current is falthy'
    }
  }
  const setCursor = useCallback(
    (insertIndex: number) => {
      if (inputRef.current) {
        inputRef.current.focus()
        inputRef.current.setSelectionRange(insertIndex, insertIndex)
      }
    },
    [inputRef]
  )
  return { insert }
}

export default useInsertQuery
