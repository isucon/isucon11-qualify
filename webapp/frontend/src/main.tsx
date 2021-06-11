import { StrictMode } from 'react'
import ReactDOM from 'react-dom'
import './index.css'
import 'windi.css'
import App from './App'
import { BrowserRouter, Route, Switch } from 'react-router-dom'
import GuardedRoute from './router/GuardedRoute'
import Auth from './pages/Auth'

ReactDOM.render(
  <StrictMode>
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
  </StrictMode>,
  document.getElementById('root')
)
