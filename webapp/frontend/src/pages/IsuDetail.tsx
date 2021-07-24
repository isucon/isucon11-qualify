import MainInfo from '../components/IsuDetail/MainInfo'
import Card from '../components/UI/Card'
import NowLoading from '../components/UI/NowLoading'
import { Isu } from '../lib/apis'

interface Props {
  isu: Isu
  setIsu: React.Dispatch<React.SetStateAction<Isu | null>>
}

const IsuDetail = ({ isu, setIsu }: Props) => {
  if (!isu) {
    return <NowLoading />
  }
  return (
    <div className="flex flex-col gap-10 items-center">
      <Card>
        <MainInfo isu={isu} setIsu={setIsu} />
      </Card>
    </div>
  )
}

export default IsuDetail
