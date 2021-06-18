import { StrictMode } from 'react'
import ReactDOM from 'react-dom'
import './index.css'
import 'windi.css'
import App from './App'
import StateContextProvider from './context/state'

ReactDOM.render(
  <StrictMode>
    <StateContextProvider>
      <App />
    </StateContextProvider>
  </StrictMode>,
  document.getElementById('root')
)
