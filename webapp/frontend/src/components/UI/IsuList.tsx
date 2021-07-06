import { Isu as IIsu } from '../../lib/apis'
import Isu from './Isu'

interface Props {
  isus: IIsu[]
}

const IsuList = ({ isus }: Props) => {
  return (
    <div className="grid gap-8 grid-cols-isus w-full">
      {isus.map(isu => (
        <Isu isu={isu} key={isu.jia_isu_uuid} />
      ))}
    </div>
  )
}

export default IsuList
