import { Isu } from '/@/lib/apis'
import Tabs from './Tabs'

const SubHeader = ({ isu }: { isu: Isu }) => {
  return (
    <header className="bg-secondary pl-16 pt-8">
      <h2 className="mb-8 ml-4 text-2xl font-bold">{isu.name}</h2>
      <Tabs id={isu.jia_isu_uuid} />
    </header>
  )
}

export default SubHeader
