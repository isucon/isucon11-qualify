import Card from '../components/UI/Card'
import NowLoading from '../components/UI/NowLoading'
import { Isu } from '../lib/apis'

interface Props {
  isu: Isu
  setIsu: React.Dispatch<React.SetStateAction<Isu | null>>
}

const IsuGraph = ({ isu, setIsu }: Props) => {
  if (!isu) {
    return <NowLoading />
  }
  return <div>グラフページ</div>
}

export default IsuGraph
