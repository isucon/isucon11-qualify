import { useEffect } from 'react'
import { useState } from 'react'
import { useParams } from 'react-router-dom'
import CatalogInfo from '../components/IsuDetail/Catalog'
import NowLoading from '../components/UI/NowLoading'
import apis, { Isu } from '../lib/apis'

const IsuDetail = () => {
  const [isu, setIsu] = useState<Isu | null>(null)
  const { id } = useParams<{ id: string }>()

  useEffect(() => {
    const load = async () => {
      setIsu(await apis.getIsu(id))
    }
    load()
  }, [id])

  if (!isu) {
    return <NowLoading />
  }
  return (
    <div>
      <div>椅子詳細</div>
      <div>{isu.name}</div>
      <CatalogInfo isu={isu} />
    </div>
  )
}

export default IsuDetail
