import React, { useEffect, useRef, useState } from 'react'
import type { Client } from '~/api'
import { type GraphPoints, RoutesGraph } from './components'

export default function Containers({ apiClient }: { apiClient: Client }): React.JSX.Element {
  const [routes, setRoutes] = useState<ReadonlyMap<string, ReadonlyArray<URL>> | null>(null)
  const [isLoading, setIsLoading] = useState<boolean>(false)
  const [graphPoints, setGraphPoints] = useState<GraphPoints | null>(null)
  const closeRoutesSub = useRef<(() => void) | null>(null)

  // fetch the list of routes and subscribe to updates
  useEffect(() => {
    setIsLoading(true)

    apiClient
      .routesList()
      .then((routes): void => setRoutes(routes))
      .catch(console.error)
      .finally(() => setIsLoading(false))

    apiClient
      .routesSubscribe((routes): void => setRoutes(routes))
      .then((closer): void => {
        closeRoutesSub.current = closer // save the closer function to call it when the component unmounts
      })
      .catch(console.error)

    return () => {
      // close the WebSocket connection when the component unmounts
      if (closeRoutesSub.current) {
        closeRoutesSub.current()
      }
    }
  }, [apiClient])

  // on routes update, update the graph points
  useEffect(() => {
    const root = 'indocker.app'
    const points: GraphPoints = new Map()
    const { protocol, port } = window.location

    for (const builtin of [root, `monitor.${root}`]) {
      points.set(builtin, { isBuiltIn: true, url: new URL(`${protocol}//${builtin}` + (port ? `:${port}` : '')) })
    }

    if (routes) {
      for (const [hostname, urls] of routes) {
        const domain = `${hostname}.${root}`

        points.set(domain, { url: urls.length ? new URL(`${protocol}//${domain}` + (port ? `:${port}` : '')) : null })
      }
    }

    setGraphPoints(points)
  }, [routes])

  return (
    <>
      <h1>Containers</h1>
      <RoutesGraph loading={isLoading} points={graphPoints} />
    </>
  )
}