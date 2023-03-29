import type { Router } from 'vue-router'
import { useRouter } from 'vue-router'
import { createRouter, createWebHistory } from 'vue-router'
import ViewAllContainers from '@/components/containers/ViewAll.vue'
import ViewPorts from '@/components/ports/ViewAll.vue'
import ViewPreferences from '@/components/preferences/ViewAll.vue'
import ViewStatsMonitor from '@/components/stats/ViewAll.vue'
import ViewNotFound from '@/components/ViewNotFound.vue'
import ViewContainerLogs from '@/components/containers/logs/ViewLogs.vue'
import ViewContainerStats from '@/components/containers/stats/ViewStats.vue'
import type { Component } from 'vue'
import type { RouteParamsRaw, RouteRecord } from 'vue-router'
import {
  Build as PreferencesIcon,
  LogoDocker as DockerIcon,
  StatsChart as StatsIcon,
  SwapVertical as PortsIcon
} from '@vicons/ionicons5'

/**
 * Note: Children route names must be prefixed with the parent route name (separator: dot).
 */
export enum RouteName {
  Containers = 'containers',
  ContainerLogs = 'containers.logs',
  ContainerStats = 'containers.stats',
  Stats = 'stats',
  Ports = 'ports',
  Preferences = 'preferences',
  NotFound = 'not-found'
}

declare module 'vue-router' {
  interface RouteMeta {
    visible?: boolean // is displayed in the menu
    title?: string // title of the page
    icon?: Component
  }
}

export function router(): Router {
  return createRouter({
    history: createWebHistory(),
    routes: [
      {
        // "home" -> "containers" redirect
        path: '/',
        redirect: { name: RouteName.Containers }
      },
      {
        path: '/containers/:id?', // /containers/<ID>
        name: RouteName.Containers,
        component: ViewAllContainers,
        meta: {
          visible: true,
          title: 'Containers',
          icon: DockerIcon
        },
        children: [
          {
            path: 'logs', // /containers/<ID>/logs
            name: RouteName.ContainerLogs,
            component: ViewContainerLogs
          },
          {
            path: 'stats', // /containers/<ID>/stats
            name: RouteName.ContainerStats,
            component: ViewContainerStats
          }
        ]
      },
      {
        // "/containers/foo-id" -> "/containers/foo-id/logs" redirect
        path: '/containers/:id',
        redirect: { name: RouteName.ContainerLogs }
      },
      {
        path: '/stats',
        name: RouteName.Stats,
        component: ViewStatsMonitor,
        meta: {
          visible: true,
          title: 'Stats monitor',
          icon: StatsIcon
        }
      },
      {
        path: '/ports',
        name: RouteName.Ports,
        component: ViewPorts,
        meta: {
          visible: true,
          title: 'Ports',
          icon: PortsIcon
        }
      },
      {
        path: '/preferences',
        name: RouteName.Preferences,
        component: ViewPreferences,
        meta: {
          visible: true,
          title: 'Preferences',
          icon: PreferencesIcon
        }
      },
      {
        path: '/:pathMatch(.*)*',
        name: RouteName.NotFound,
        component: ViewNotFound
      }
    ]
  })
}

/** Get the container ID from current route. */
export function id(): string | undefined {
  const currentParams = useRouter().currentRoute.value.params

  if (Object.prototype.hasOwnProperty.call(currentParams, 'id')) {
    return currentParams.id as string
  }
}

/** Get the visible routes. */
export function visible(): RouteRecord[] {
  return useRouter().getRoutes().filter((route) => route.meta.visible)
}

/** Go to a route. */
export function goto(router: Router, name: RouteName, params?: RouteParamsRaw): void {
  router.push({ name: name, params: params })
}

/** Get the current route name. If `parentOnly` is true, only the parent route name is returned. */
export function current(parentOnly?: boolean): RouteName | undefined {
  const current = useRouter().currentRoute?.value?.name as RouteName

  if (typeof current === 'string') {
    if (parentOnly) {
      return current.split('.')[0] as RouteName
    }

    return current
  }
}
