import React, { useEffect, useRef, useState } from 'react'
import type { Client } from '~/api'
import { AnimatedLayout } from '~/shared/components'
import { type ContainerListItem, ContainersList } from './components'
import styles from './screen.module.scss'

export default function Screen({ apiClient }: { apiClient: Client }): React.JSX.Element {
  const [routes, setRoutes] = useState<ReadonlyMap<string, ReadonlyMap<string, URL>> | null>(null) // map<hostname, map<container_id, url>>
  const [, setIsLoading] = useState<boolean>(false)
  const [listItems, setListItems] = useState<ReadonlyArray<ContainerListItem>>([])
  const closeRoutesSub = useRef<(() => void) | null>(null)

  // fetch the list of routes and subscribe to updates
  useEffect(() => {
    Promise.resolve().then(() => setIsLoading(true))

    apiClient
      .routesList()
      .then((routes): void => setRoutes(routes))
      .catch(console.error)
      .finally(() => setIsLoading(false))

    apiClient
      .routesSubscribe({
        onUpdate: (routes): void => setRoutes(routes),
        onError: console.warn,
      })
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

  // watch for changes in the routes map and update the list of items
  useEffect(() => {
    if (!routes) {
      return
    }

    const items: Array<ContainerListItem> = []

    for (const [hostname, containers] of routes.entries()) {
      items.push({
        hostname: hostname,
        routes: containers,
        url: new URL(
          `https://${hostname}.indocker.app` + (window.location.port ? `:${window.location.port}` : '') + '/'
        ),
      })
    }

    Promise.resolve().then(() => setListItems(items))
  }, [routes])

  return (
    <AnimatedLayout>
      <div className={styles.containerOuter}>
        <div className={styles.containerInner}>
          <ContainersList apiClient={apiClient} items={listItems} />
        </div>
        <div className={styles.waveWithBubbles} />
        {/*<RoutesGraph loading={isLoading} points={graphPoints} />*/}
      </div>
    </AnimatedLayout>
  )
}
