import { useEffect } from 'react'
import { useState } from 'react'
import apis, { Catalog, Isu } from '../../lib/apis'
import NowLoading from '../UI/NowLoading'

interface Props {
  isu: Isu
}

const CatalogInfo = ({ isu }: Props) => {
  const [catalog, setCatalog] = useState<Catalog | null>(null)

  useEffect(() => {
    const fetchCatalog = async () => {
      setCatalog(await apis.getCatalog(isu.jia_catalog_id))
    }
    fetchCatalog()
  }, [isu])

  if (!catalog) {
    return <NowLoading />
  }
  return <div>{JSON.stringify(catalog)}</div>
}

export default CatalogInfo
