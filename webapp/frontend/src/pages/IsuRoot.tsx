import { useEffect, useState } from 'react'
import { Switch, useParams } from 'react-router-dom'
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

  return (
    <div>
      <div>tab</div>
      <Switch>{isu ? <DefineRoutes isu={isu} /> : <NowLoading />}</Switch>
    </div>
  )
}

const DefineRoutes = ({ isu }: { isu: Isu }) => (
  <GuardedRoute path="/isu/:id" exact>
    <IsuDetail isu={isu} />
  </GuardedRoute>
)

export default IsuRoot
