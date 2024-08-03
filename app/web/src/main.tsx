import React from 'react'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { apiClient } from '~/api'
import '~/theme/app.scss'

const App = (): React.JSX.Element => {
  queueMicrotask(async () => {
    await Promise.all([apiClient.currentVersion(), apiClient.latestVersion()]).then(([version, latest]) => {
      console.log(`Version: ${version}, Latest: ${latest}`)
    })
  })

  return <p>Hello world!</p>
}

createRoot(document.getElementById('root') as HTMLElement).render(
  <StrictMode>
    <App />
  </StrictMode>
)
