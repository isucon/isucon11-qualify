import { BrowserRouter, Switch, Route } from 'react-router-dom'
import PageHeader from './components/PageHeader/PageHeader'
import Auth from './pages/Auth'
import Home from './pages/Home'
import IsuRoot from './pages/IsuRoot'
import Register from './pages/Register'
import GuardedRoute from './router/GuardedRoute'

const App = () => {
  return (
    <div className="flex flex-col min-h-full text-primary">
      <BrowserRouter>
        <PageHeader></PageHeader>
        <div className="flex-1 bg-primary">
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
            <Route path="/login">
              <Auth />
            </Route>
          </Switch>
        </div>
      </BrowserRouter>
    </div>
  )
}

export default App
