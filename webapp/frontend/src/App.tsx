import { BrowserRouter, Switch, Route } from 'react-router-dom'
import PageHeader from './components/PageHeader/PageHeader'
import Auth from './pages/Auth'
import GuardedRoute from './router/GuardedRoute'

const App = () => {
  return (
    <div>
      <PageHeader></PageHeader>
      <BrowserRouter>
        <Switch>
          <GuardedRoute path="/" exact></GuardedRoute>
          <Route path="/login">
            <Auth />
          </Route>
        </Switch>
      </BrowserRouter>
    </div>
  )
}

export default App
