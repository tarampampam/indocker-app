import React, { useEffect, useState } from 'react'
import type { Client } from '~/api'
import { RoutesGraph, type GraphPoints } from './components'

export default function Containers({ apiClient }: { apiClient: Client }): React.JSX.Element {
  // const [routes, setRoutes] = useState<ReadonlyMap<string, ReadonlyArray<URL>> | null>(null)
  const [isLoading, setIsLoading] = useState<boolean>(false)
  const [graphPoints, setGraphPoints] = useState<GraphPoints | null>(null)

  useEffect(() => {
    setIsLoading(true)

    apiClient
      .routesList()
      .then((routes): void => {
        // setRoutes(routes)

        const root = 'indocker.app'
        const points: GraphPoints = new Map()
        const { protocol, port } = window.location

        for (const builtin of [root, `monitor.${root}`]) {
          points.set(builtin, { isBuiltIn: true, url: new URL(`${protocol}//${builtin}` + (port ? `:${port}` : '')) })
        }

        for (const [hostname, urls] of routes) {
          const domain = `${hostname}.${root}`

          points.set(domain, { url: urls.length ? new URL(`${protocol}//${domain}` + (port ? `:${port}` : '')) : null })
        }

        console.log(points)

        setGraphPoints(points)
      })
      .catch(console.error)
      .finally(() => setIsLoading(false))
  }, [apiClient])

  return (
    <>
      <h1>Containers</h1>
      <RoutesGraph loading={isLoading} points={graphPoints} />
    </>
  )
}
