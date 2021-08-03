import MainInfo from '../components/IsuDetail/MainInfo'
import Card from '../components/UI/Card'
import NowLoading from '../components/UI/NowLoading'
import { Isu } from '../lib/apis'

interface Props {
  isu: Isu
}

const IsuDetail = ({ isu }: Props) => {
  if (!isu) {
    return <NowLoading />
  }
  return (
    <div className="flex flex-col gap-10 items-center">
      <Card>
        <MainInfo isu={isu} />
      </Card>
    </div>
  )
}

export default IsuDetail
