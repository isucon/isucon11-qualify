import { BrowserRouter, Redirect, Route, Switch } from 'react-router-dom'
import { Toaster } from 'react-hot-toast'
import PageHeader from './components/PageHeader/PageHeader'
import Home from './pages/Home'
import IsuRoot from './pages/IsuRoot'
import Register from './pages/Register'
import GuardedRoute from './router/GuardedRoute'
import ReactTooltip from 'react-tooltip'

const App = () => {
  return (
    <div className="text-primary flex flex-col min-h-full">
      <Toaster position="bottom-left" />
      <ReactTooltip />
      <BrowserRouter>
        <PageHeader></PageHeader>
        <div className="bg-primary relative flex-grow">
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
