import { useState, useEffect } from 'react'
import Card from '../components/UI/Card'
import IsuList from '../components/UI/IsuList'
import apis, { Isu } from '../lib/apis'
import useSearch from '../components/Search/use/search'
import SearchInput from '../components/Search/SearchInput'
import PagingNavigator from '../components/UI/PagingNavigator'

const Search = () => {
  const [isus, setIsus] = useState<Isu[]>([])
  useEffect(() => {
    const fetchIsus = async () => {
      setIsus(await apis.getIsuSearch())
    }
    fetchIsus()
  }, [setIsus])

  const { query, page } = useSearch()

  return (
    <div className="flex justify-center p-10">
      <Card>
        <div className="flex flex-col items-center">
          <h2 className="text-xl font-bold">ISU</h2>
          <SearchInput
            query={query}
            search={() => new Promise(() => undefined)}
          />
          <IsuList isus={isus} />
          <PagingNavigator
            prev={() => undefined}
            next={() => undefined}
            length={isus.length}
            maxLength={10}
            page={page}
          />
        </div>
      </Card>
    </div>
  )
}

export default Search
