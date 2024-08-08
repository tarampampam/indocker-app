import React from 'react'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import { routes } from './router'
import { echarts as echartsThemes } from './theme'
import { registerTheme as echartsRegisterTheme } from 'echarts/core'
import '~/theme/app.scss'

echartsRegisterTheme('dark', echartsThemes.dark)
echartsRegisterTheme('light', echartsThemes.light)

const App = (): React.JSX.Element => {
  return <RouterProvider router={createBrowserRouter(routes)} />
}

createRoot(document.getElementById('root') as HTMLElement).render(
  <StrictMode>
    <App />
  </StrictMode>
)
