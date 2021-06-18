import { BrowserRouter, Switch, Route } from 'react-router-dom'
import PageHeader from './components/PageHeader/PageHeader'
import Auth from './pages/Auth'
import GuardedRoute from './router/GuardedRoute'

const App = () => {
  return (
    <div className="text-primary-700">
      <BrowserRouter>
        <PageHeader></PageHeader>
        <Switch>
          <GuardedRoute path="/" exact></GuardedRoute>
          <GuardedRoute path="/isu/condition" exact>
            <div>通知画面</div>
          </GuardedRoute>
          <GuardedRoute path="/isu/search" exact>
            <div>検索画面</div>
          </GuardedRoute>
          <Route path="/login">
            <Auth />
          </Route>
        </Switch>
      </BrowserRouter>
    </div>
  )
}

export default App
