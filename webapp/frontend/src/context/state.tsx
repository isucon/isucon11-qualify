import { createContext, Dispatch, useContext, useReducer } from 'react'
import { User } from '/@/lib/apis'

export interface State {
  me: User | null
}
type Action = { type: 'login'; user: User } | { type: 'logout' }

const initialState: State = {
  me: null
}

const stateContext = createContext<State>(initialState)
const dispatchContext = createContext<Dispatch<Action> | null>(null)

const stateReducer = (state: State, action: Action): State => {
  switch (action.type) {
    case 'login': {
      return { ...state, me: action.user }
    }
    case 'logout': {
      return { ...state, me: null }
    }
    default: {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const _exhaustiveCheck: never = action
      throw new Error('invalid action')
    }
  }
}

const StateContextProvider = ({ children }: { children: JSX.Element }) => {
  const [state, dispatch] = useReducer(stateReducer, initialState)

  return (
    <stateContext.Provider value={state}>
      <dispatchContext.Provider value={dispatch}>
        {children}
      </dispatchContext.Provider>
    </stateContext.Provider>
  )
}

export const useStateContext = () => {
  const state = useContext(stateContext)
  if (!state) {
    throw new Error('require wrapped by stateContext')
  }
  return state
}

export const useDispatchContext = () => {
  const dispatch = useContext(dispatchContext)
  if (!dispatch) {
    throw new Error('require wrapped by stateContext')
  }
  return dispatch
}

export default StateContextProvider
