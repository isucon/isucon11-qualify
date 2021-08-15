import { Link } from 'react-router-dom'
import { Isu as IIsu } from '/@/lib/apis'

interface Props {
  isu: IIsu
}

const Isu = ({ isu }: Props) => {
  return (
    <Link
      to={`/isu/${isu.jia_isu_uuid}`}
      className="flex flex-col items-center"
    >
      <img
        src={`/api/isu/${isu.jia_isu_uuid}/icon`}
        className="w-48 h-48 object-contain"
        key={isu.jia_isu_uuid}
      />
      <h3>{isu.name}</h3>
    </Link>
  )
}

export default Isu
