import React from 'react'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { apiClient } from '~/api'
import '~/theme/app.scss'

const App = (): React.JSX.Element => {
  queueMicrotask(async () => {
    await Promise.all([
      apiClient.currentVersion(),
      apiClient.latestVersion(),
      apiClient.routesList(),
    ]).then(([version, latest, routes]) => {
      console.log(`version: ${version}`)
      console.log(`latest: ${latest}`)
      console.log('routes: ', routes)
    })
  })

  return <p>Hello world!1!</p>
}

createRoot(document.getElementById('root') as HTMLElement).render(
  <StrictMode>
    <App />
  </StrictMode>
)
