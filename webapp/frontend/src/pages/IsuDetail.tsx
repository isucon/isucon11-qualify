import { Link } from 'react-router-dom'
import CatalogInfo from '../components/IsuDetail/Catalog'
import MainInfo from '../components/IsuDetail/MainInfo'
import Card from '../components/UI/Card'
import NowLoading from '../components/UI/NowLoading'
import { Isu } from '../lib/apis'

const IsuDetail = ({ isu }: { isu: Isu }) => {
  if (!isu) {
    return <NowLoading />
  }
  return (
    <div className="flex flex-col items-center">
      <Card>
        <MainInfo isu={isu} />
      </Card>
      <div>椅子詳細</div>
      <CatalogInfo isu={isu} />
      <Link to={`/isu/${isu.jia_isu_uuid}/graph`}>グラフの確認</Link>
      <Link to={`/isu/${isu.jia_isu_uuid}/condition`}>状態の確認</Link>
    </div>
  )
}

export default IsuDetail
