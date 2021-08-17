import { Isu } from '/@/lib/apis'

interface Props {
  isu: Isu
  customClass: string
}

const IsuImage = ({ isu, customClass }: Props) => {
  return (
    <img
      src={`/api/isu/${isu.jia_isu_uuid}/icon`}
      className={`rounded object-cover ` + customClass}
      key={isu.jia_isu_uuid}
    />
  )
}

export default IsuImage
