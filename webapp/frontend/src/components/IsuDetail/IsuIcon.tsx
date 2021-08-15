import { Isu } from '/@/lib/apis'

interface Props {
  isu: Isu
  reloadKey?: number
}
const IsuIcon = ({ isu, reloadKey }: Props) => {
  return (
    <img
      src={`/api/isu/${isu.jia_isu_uuid}/icon`}
      className="w-48 h-48 object-contain"
      key={reloadKey}
    />
  )
}

export default IsuIcon
