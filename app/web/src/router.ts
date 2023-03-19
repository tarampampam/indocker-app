import type { Router } from 'vue-router'
import { createRouter, createWebHistory } from 'vue-router'
import ViewContainers from '@/components/ViewContainers.vue'
import ViewPorts from '@/components/ViewPorts.vue'
import ViewPreferences from '@/components/ViewPreferences.vue'
import ViewStatsMonitor from '@/components/ViewStatsMonitor.vue'
import ViewNotFound from '@/components/ViewNotFound.vue'
import ViewContainerLogs from '@/components/ViewContainerLogs.vue'
import ViewContainerStats from '@/components/ViewContainerStats.vue'
import type { Component } from 'vue'
import {
  Build as PreferencesIcon,
  LogoDocker as DockerIcon,
  StatsChart as StatsIcon,
  SwapVertical as PortsIcon
} from '@vicons/ionicons5'

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
        redirect: { name: 'containers' }
      },
      {
        path: '/containers/:id?',
        name: 'containers',
        component: ViewContainers,
        meta: {
          visible: true,
          title: 'Containers',
          icon: DockerIcon
        },
        children: [
          {
            path: 'logs',
            name: 'containers.logs',
            component: ViewContainerLogs
          },
          {
            path: 'stats',
            name: 'containers.stats',
            component: ViewContainerStats
          }
        ]
      },
      {
        // "/containers/foo-id" -> "/containers/foo-id/logs" redirect
        path: '/containers/:id',
        redirect: { name: 'containers.logs' }
      },
      {
        path: '/stats',
        name: 'stats',
        component: ViewStatsMonitor,
        meta: {
          visible: true,
          title: 'Stats monitor',
          icon: StatsIcon
        }
      },
      {
        path: '/ports',
        name: 'ports',
        component: ViewPorts,
        meta: {
          visible: true,
          title: 'Ports',
          icon: PortsIcon
        }
      },
      {
        path: '/preferences',
        name: 'preferences',
        component: ViewPreferences,
        meta: {
          visible: true,
          title: 'Preferences',
          icon: PreferencesIcon
        }
      },
      {
        path: '/:pathMatch(.*)*',
        name: 'not-found',
        component: ViewNotFound
      }
    ]
  })
}
