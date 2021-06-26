import { useEffect } from 'react'
import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import CatalogInfo from '../components/IsuDetail/Catalog'
import IsuIcon from '../components/IsuDetail/IsuIcon'
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
      <IsuIcon isu={isu} />
      <Link to={`/isu/${isu.jia_isu_uuid}/graph`}>グラフの確認</Link>
      <Link to={`/isu/${isu.jia_isu_uuid}/condition`}>状態の確認</Link>
    </div>
  )
}

export default IsuDetail
