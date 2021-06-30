import { useEffect, useState } from 'react'
import apis, { Isu } from '../../../lib/apis'
import { getRequestParams, validateParams } from './parseQuery'

const useSearch = () => {
  const [query, setQuery] = useState('')
  const [page, setPage] = useState(1)
  const [isus, setIsus] = useState<Isu[]>([])
  useEffect(() => {
    const fetchIsus = async () => {
      setIsus(await apis.getIsuSearch())
    }
    fetchIsus()
  }, [setIsus])

  const search = async (newQuery: string) => {
    const params = getRequestParams(newQuery)
    try {
      validateParams(params)
      params.page = '1'
      setIsus(await apis.getIsuSearch(params))
      setQuery(query)
      setPage(1)
    } catch (e) {
      alert(e)
    }
  }
  const next = async () => {
    const params = getRequestParams(query)
    params.page = `${page + 1}`
    setIsus(await apis.getIsuSearch(params))
    setPage(page + 1)
  }
  const prev = async () => {
    const params = getRequestParams(query)
    params.page = `${page - 1}`
    setIsus(await apis.getIsuSearch(params))
    setPage(page - 1)
  }
  return { query, search, isus, next, prev, page }
}

export default useSearch
