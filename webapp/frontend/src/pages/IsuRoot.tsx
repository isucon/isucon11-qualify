import { useEffect, useState } from 'react'
import { Switch, useParams } from 'react-router-dom'
import SubHeader from '../components/Isu/SubHeader'
import NowLoading from '../components/UI/NowLoading'
import apis, { Isu } from '../lib/apis'
import GuardedRoute from '../router/GuardedRoute'
import IsuCondition from './IsuCondition'
import IsuDetail from './IsuDetail'
import IsuGraph from './IsuGraph'

const IsuRoot = () => {
  const [isu, setIsu] = useState<Isu | null>(null)
  const { id } = useParams<{ id: string }>()

  useEffect(() => {
    const load = async () => {
      setIsu(await apis.getIsu(id))
    }
    load()
  }, [id])

  if (!isu) {
    return <NowLoading />
  }
  return (
    <div>
      <SubHeader isu={isu} />
      <div className="p-10">
        <DefineRoutes isu={isu} setIsu={setIsu} />
      </div>
    </div>
  )
}

const DefineRoutes = ({
  isu,
  setIsu
}: {
  isu: Isu
  setIsu: React.Dispatch<React.SetStateAction<Isu | null>>
}) => (
  <Switch>
    <GuardedRoute path="/isu/:id" exact>
      <IsuDetail isu={isu} setIsu={setIsu} />
    </GuardedRoute>
    <GuardedRoute path="/isu/:id/condition" exact>
      <IsuCondition isu={isu} />
    </GuardedRoute>
    <GuardedRoute path="/isu/:id/graph" exact>
      <IsuGraph isu={isu} />
    </GuardedRoute>
  </Switch>
)

export default IsuRoot
