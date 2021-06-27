import { useState } from 'react'
import { Isu } from '../../lib/apis'
import IconInput from '../UI/IconInput'

const IsuIcon = ({ isu }: { isu: Isu }) => {
  const [key, setKey] = useState(0)
  const reloadIcon = () => {
    // 画像をアップデートしてもsrcとかが更新されるわけではないので再レンダリングのためセットしている
    setKey(performance.now())
  }

  return (
    <div>
      <img src={`/api/isu/${isu.jia_isu_uuid}/icon`} key={key} />
      <IconInput isu={isu} reloadIcon={reloadIcon} />
    </div>
  )
}

export default IsuIcon
