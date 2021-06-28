import { useEffect, useState } from 'react'
import { Switch, useParams } from 'react-router-dom'
import SubHeader from '../components/Isu/SubHeader'
import NowLoading from '../components/UI/NowLoading'
import apis, { Isu } from '../lib/apis'
import GuardedRoute from '../router/GuardedRoute'
import IsuDetail from './IsuDetail'

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
      <Switch>
        <DefineRoutes isu={isu} />
      </Switch>
    </div>
  )
}

const DefineRoutes = ({ isu }: { isu: Isu }) => (
  <GuardedRoute path="/isu/:id" exact>
    <IsuDetail isu={isu} />
  </GuardedRoute>
)

export default IsuRoot
