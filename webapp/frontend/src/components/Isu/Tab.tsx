import { Link, useLocation } from 'react-router-dom'

const Tab = () => {
  const { pathname } = useLocation()

  const getIsuId = () => pathname.split('/')[2]
  return (
    <div>
      <Link to={`/isu/${getIsuId()}`}>詳細</Link>
      <Link to={`/isu/${getIsuId()}/condition`}>状態</Link>
      <Link to={`/isu/${getIsuId()}/graph`}>グラフ</Link>
    </div>
  )
}

export default Tab
