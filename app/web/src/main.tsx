import React, { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import { routes } from './routing'
import '~/theme/app.scss'

const App = (): React.JSX.Element => {
  return <RouterProvider router={createBrowserRouter(routes)} />
}

const registerServiceWorker = async (): Promise<ServiceWorkerRegistration> => {
  if (location.protocol !== 'https:') {
    return Promise.reject(new Error('Service workers are only supported over HTTPS'))
  }

  if (!('serviceWorker' in navigator)) {
    return Promise.reject(new Error('Service workers are not supported'))
  }

  const reg = await navigator.serviceWorker.register('./service-worker.js', { scope: '/', type: 'module' })

  reg.addEventListener('updatefound', () => {
    const newWorker = reg.installing

    newWorker?.addEventListener('statechange', () => {
      if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
        console.debug('an update for the service worker is available; reload the page to update')
      }
    })
  })

  return reg
}

createRoot(document.getElementById('root') as HTMLElement).render(
  <StrictMode>
    <App />
  </StrictMode>,
)

registerServiceWorker().catch(console.warn)
