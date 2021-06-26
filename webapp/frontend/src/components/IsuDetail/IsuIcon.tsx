import { Isu } from '../../lib/apis'
import IconInput from '../UI/IconInput'

const IsuIcon = ({ isu }: { isu: Isu }) => {
  return (
    <div>
      <img src={`/api/isu/${isu.jia_isu_uuid}/icon`} />
      <IconInput isu={isu} />
    </div>
  )
}

export default IsuIcon
