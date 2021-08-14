import { useEffect, useState } from 'react'
import { Redirect, Switch, useParams } from 'react-router-dom'
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
  const [notFound, setNotFound] = useState(false)

  useEffect(() => {
    const load = async () => {
      try {
        setIsu(await apis.getIsu(id))
      } catch (e) {
        if (e.response.status === 404) {
          setNotFound(true)
        }
      }
    }
    load()
  }, [id, setNotFound])

  if (!isu) {
    if (notFound) {
      return <Redirect to={`/`} />
    }
    return <NowLoading />
  }

  return (
    <div>
      <SubHeader isu={isu} />
      <div className="p-10">
        <DefineRoutes isu={isu} />
      </div>
    </div>
  )
}

const DefineRoutes = ({ isu }: { isu: Isu }) => (
  <Switch>
    <GuardedRoute path="/isu/:id" exact>
      <IsuDetail isu={isu} />
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
