import { Isu } from '../../lib/apis'

interface Props {
  isu: Isu
  reloadKey?: number
}
const IsuIcon = ({ isu, reloadKey }: Props) => {
  return (
    <div>
      <img src={`/api/isu/${isu.jia_isu_uuid}/icon`} key={reloadKey} />
    </div>
  )
}

export default IsuIcon
