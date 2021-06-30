import Card from '../components/UI/Card'
import IsuList from '../components/UI/IsuList'
import { DEFAULT_SEARCH_LIMIT } from '../lib/apis'
import useSearch from '../components/Search/use/search'
import SearchInput from '../components/Search/SearchInput'
import PagingNavigator from '../components/UI/PagingNavigator'

const Search = () => {
  const { query, page, isus, prev, next, search } = useSearch()

  return (
    <div className="flex justify-center p-10">
      <Card>
        <div className="flex flex-col items-center">
          <h2 className="text-xl font-bold">ISU</h2>
          <SearchInput query={query} search={search} />
          <IsuList isus={isus} />
          <PagingNavigator
            prev={prev}
            next={next}
            length={isus.length}
            maxLength={DEFAULT_SEARCH_LIMIT}
            page={page}
          />
        </div>
      </Card>
    </div>
  )
}

export default Search
