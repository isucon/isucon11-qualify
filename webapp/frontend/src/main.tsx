import { StrictMode } from 'react'
import ReactDOM from 'react-dom'
import './index.css'
import 'windi.css'
import App from './App'
import { BrowserRouter, Route, Switch } from 'react-router-dom'
import GuardedRoute from './router/GuardedRoute'
import Auth from './pages/Auth'
import StateContextProvider from './context/state'

ReactDOM.render(
  <StrictMode>
    <StateContextProvider>
      <BrowserRouter>
        <Switch>
          <GuardedRoute path="/" exact>
            <App />
          </GuardedRoute>
          <Route path="/login">
            <Auth />
          </Route>
        </Switch>
      </BrowserRouter>
    </StateContextProvider>
  </StrictMode>,
  document.getElementById('root')
)
