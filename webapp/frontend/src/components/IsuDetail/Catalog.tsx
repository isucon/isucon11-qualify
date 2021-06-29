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
  return (
    <div>
      <div className="mb-3 text-xl font-bold">Catalog</div>
      <div className="flex flex-col gap-2 pl-2">
        <CatalogRow property="name" value={catalog.name} />
        <CatalogRow property="size" value={catalog.size} />
        <CatalogRow property="weight" value={`${catalog.weight}`} />
        <CatalogRow property="limit weight" value={`${catalog.limit_weight}`} />
        <CatalogRow property="maker" value={catalog.maker} />
        <CatalogRow property="tags" value={catalog.tags} />
      </div>
    </div>
  )
}

const CatalogRow = ({
  property,
  value
}: {
  property: string
  value: string
}) => {
  return (
    <div className="border-b-1 flex pl-1 border-outline">
      <div className="flex-1">{property}</div>
      <div className="flex-1">{value}</div>
    </div>
  )
}

export default CatalogInfo
