import { Navigate, matchRoutes, useLocation, type RouteObject } from 'react-router-dom'
import { default as DefaultLayout } from '~/screens/layout'
import { NotFoundScreen } from '~/screens/not-found'
import { ContainersScreen } from '~/screens/containers'
import { AboutScreen } from '~/screens/about'
import { apiClient } from '~/api'

export enum RouteIDs {
  Home = 'home',
  Containers = 'containers',
  About = 'about',
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
        element: <Navigate to={pathTo(RouteIDs.Containers)} />,
      },
      {
        path: 'containers',
        id: RouteIDs.Containers,
        element: <ContainersScreen apiClient={apiClient} />,
      },
      {
        path: 'about',
        id: RouteIDs.About,
        element: <AboutScreen />,
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
    case RouteIDs.Containers:
      return '/containers'
    case RouteIDs.About:
      return '/about'
    default:
      throw new Error(`Unknown route: ${path}`)
  }
}
