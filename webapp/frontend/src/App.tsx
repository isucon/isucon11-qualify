import { BrowserRouter, Redirect, Route, Switch } from 'react-router-dom'
import PageHeader from './components/PageHeader/PageHeader'
import Home from './pages/Home'
import IsuRoot from './pages/IsuRoot'
import Register from './pages/Register'
import GuardedRoute from './router/GuardedRoute'

const App = () => {
  return (
    <div className="text-primary flex flex-col min-h-full">
      <BrowserRouter>
        <PageHeader></PageHeader>
        <div className="bg-primary flex-1">
          <Switch>
            <GuardedRoute path="/" exact>
              <Home />
            </GuardedRoute>
            <GuardedRoute path="/isu/:id">
              <IsuRoot />
            </GuardedRoute>
            <GuardedRoute path="/register" exact>
              <Register />
            </GuardedRoute>
            <Route>
              <Redirect to="/" />
            </Route>
          </Switch>
        </div>
      </BrowserRouter>
    </div>
  )
}

export default App
