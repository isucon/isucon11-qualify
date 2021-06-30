import { useState } from 'react'

const useSearch = () => {
  const [query, setQuery] = useState('')
  const [page, usePage] = useState(1)

  return { query, setQuery, page }
}

export default useSearch
