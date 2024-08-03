import React from 'react'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '~/theme/app.scss'

const App = (): React.JSX.Element => {
  return <p>Hello world!</p>
}

createRoot(document.getElementById('root') as HTMLElement).render(
  <StrictMode>
    <App />
  </StrictMode>
)
