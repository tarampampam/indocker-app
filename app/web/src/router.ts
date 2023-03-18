import type { Router } from 'vue-router'
import { createRouter, createWebHistory } from 'vue-router'
import ViewContainers from '@/components/ViewContainers.vue'
import ViewPorts from '@/components/ViewPorts.vue'
import ViewPreferences from '@/components/ViewPreferences.vue'
import ViewStatsMonitor from '@/components/ViewStatsMonitor.vue'
import ViewNotFound from '@/components/ViewNotFound.vue'

declare module 'vue-router' {
  interface RouteMeta {
    visible?: boolean // is displayed in the menu
    title?: string // title of the page
  }
}

export function router(): Router {
  return createRouter({
    history: createWebHistory(),
    routes: [
      {
        path: '/',
        name: 'containers',
        component: ViewContainers,
        meta: {
          visible: true,
          title: 'Containers'
        }
      },
      {
        path: '/stats',
        name: 'stats',
        component: ViewStatsMonitor,
        meta: {
          visible: true,
          title: 'Stats monitor'
        }
      },
      {
        path: '/ports',
        name: 'ports',
        component: ViewPorts,
        meta: {
          visible: true,
          title: 'Ports'
        }
      },
      {
        path: '/preferences',
        name: 'preferences',
        component: ViewPreferences,
        meta: {
          visible: true,
          title: 'Preferences'
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
