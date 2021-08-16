import { Isu } from '/@/lib/apis'
import Tab from './Tab'

const SubHeader = ({ isu }: { isu: Isu }) => {
  return (
    <header className="bg-secondary pb-0 pl-12 pt-8">
      <h2 className="mb-3 text-2xl font-bold">{isu.name}</h2>
      <div className="ml-6">
        <Tab id={isu.jia_isu_uuid} />
      </div>
    </header>
  )
}

export default SubHeader
