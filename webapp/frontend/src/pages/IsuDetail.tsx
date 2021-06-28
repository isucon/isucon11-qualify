import { Link, useHistory } from 'react-router-dom'
import CatalogInfo from '../components/IsuDetail/Catalog'
import IsuIcon from '../components/IsuDetail/IsuIcon'
import NowLoading from '../components/UI/NowLoading'
import apis, { Isu } from '../lib/apis'

const IsuDetail = ({ isu }: { isu: Isu }) => {
  const history = useHistory()

  const deleteIsu = async () => {
    if (isu && confirm(`本当に${isu.name}を削除しますか？`)) {
      await apis.deleteIsu(isu.jia_isu_uuid)
      history.push('/')
    }
  }

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
      <button onClick={deleteIsu}>削除</button>
    </div>
  )
}

export default IsuDetail
