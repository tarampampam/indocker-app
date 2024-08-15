import React from 'react'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import { routes } from './routing'
import '~/theme/app.scss'

const App = (): React.JSX.Element => {
  return <RouterProvider router={createBrowserRouter(routes)} />
}

createRoot(document.getElementById('root') as HTMLElement).render(
  <StrictMode>
    <App />
  </StrictMode>
)

if ('serviceWorker' in navigator && location.protocol === 'https:') {
  navigator.serviceWorker.register('./service-worker.js', { scope: '/', type: 'module' }).catch(console.error)
}
