import { Isu } from '../../lib/apis'
import Tab from './Tab'

const SubHeader = ({ isu }: { isu: Isu }) => {
  return (
    <header className="p-8 pb-0 pt-4 bg-secondary">
      <h2 className="mb-3 text-xl font-bold">{isu.name}</h2>
      <Tab id={isu.jia_isu_uuid} />
    </header>
  )
}

export default SubHeader
