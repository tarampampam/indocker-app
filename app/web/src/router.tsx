import { matchRoutes, type RouteObject, useLocation } from 'react-router-dom'
import { default as DefaultLayout } from '~/screens/layout'
import { default as NotFoundScreen } from '~/screens/not-found/screen'
import { default as Containers } from '~/screens/containers/screen'
import { apiClient } from '~/api'

export enum RouteIDs {
  Home = 'home',
}

export const routes: RouteObject[] = [
  {
    path: '/',
    element: <DefaultLayout apiClient={apiClient} />,
    errorElement: <NotFoundScreen />,
    children: [
      {
        index: true,
        id: RouteIDs.Home,
        element: <Containers apiClient={apiClient} />,
      },
    ],
  },
]

/** Resolves the current route ID from the router. */
export function useCurrentRouteID(): RouteIDs | undefined {
  const match = matchRoutes(routes, useLocation())

  if (match) {
    const ids = Object.values<string>(RouteIDs)

    for (const route of match.reverse()) {
      if (route.route.id && ids.includes(route.route.id)) {
        return route.route.id as RouteIDs
      }
    }
  }

  return undefined
}

/**
 * Converts a route ID to a path to use in a link.
 *
 * @example
 * ```tsx
 * <Link to={pathTo(RouteIDs.Home)}>Go to home</Link>
 * ```
 */
export function pathTo(path: RouteIDs): string {
  switch (path) {
    case RouteIDs.Home:
      return '/'
    default:
      throw new Error(`Unknown route: ${path}`)
  }
}
