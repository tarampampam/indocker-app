<template>
  <NLayout class="container" has-sider>
    <NLayoutSider>
      <ContainersList v-if="containers.length" :containers="containers" />
      <ContainersListEmpty v-else />
    </NLayoutSider>
    <NLayoutContent content-style="padding-left: 30px">
      <div v-if="containers.length && container" class="content">
        <div class="header">
          <div class="avatar">
            <NIconWrapper :size="32" :border-radius="10" :color="stc(container?.id)">
              <NIcon :size="18" :component="DockerIcon" />
            </NIconWrapper>
          </div>
          <div class="title" :title="container?.name.length > 25 ? container?.name : undefined">
            {{ container?.name }}
          </div>
          <div class="extra">
            <ContentSwitcher />
          </div>
        </div>
        <div class="body">
          <router-view />
        </div>
        <div class="footer">
          <NBreadcrumb>
            <NBreadcrumbItem :clickable="false"> Status: Up 57 minutes (healthy) </NBreadcrumbItem>
            <NBreadcrumbItem :clickable="false"> Created: 19 March 2023</NBreadcrumbItem>
            <NBreadcrumbItem :clickable="false"> Project: App</NBreadcrumbItem>
          </NBreadcrumb>
        </div>
      </div>
      <NLayout v-else>
        <NSkeleton text :repeat="3" />
        <NSkeleton text :repeat="1" width="66%" />
      </NLayout>
    </NLayoutContent>
  </NLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import stc from 'string-to-color'
import ContainersList from './ContainersList.vue'
import ContainersListEmpty from './ContainersListEmpty.vue'
import ContentSwitcher from './ContentSwitcher.vue'
import {
  NBreadcrumb,
  NBreadcrumbItem,
  NIcon,
  NIconWrapper,
  NLayout,
  NLayoutContent,
  NLayoutSider,
  NSkeleton,
} from 'naive-ui'
import { LogoDocker as DockerIcon } from '@vicons/ionicons5'
import { RouteName } from '@/router'
import { useRouter } from 'vue-router'
import type { Container } from './types'

const router = useRouter()

const containers = ref<Container[]>([
  {
    id: 'id-1',
    name: 'app-app-1',
    tags: ['docker-compose'],
  },
  {
    id: 'id-2-foo-bar-foo-bar-foo-bar-foo-bar-foo-bar-foo-bar',
    name: 'id-2-foo-bar-foo-bar-foo-bar-foo-bar-foo-bar-foo-bar',
    tags: [
      'foo',
      'bar',
      'foo',
      'bar',
      'foofoofoofoo',
      'foo foo foo foo foo foo foo foo foo bar bar bar',
      'bar',
    ],
  },
])

onMounted((): void => {
  if (containers.value.length) {
    const currentID = router.currentRoute.value.params.id

    // redirect to first container, if no container is selected
    if (!currentID) {
      router.push({ name: RouteName.ContainerLogs, params: { id: containers.value[0].id } })

      return
    }

    // redirect to first container, if selected container does not exist
    if (!containers.value.find((c) => c.id === currentID)) {
      router.push({ name: RouteName.ContainerLogs, params: { id: containers.value[0].id } })

      return
    }
  } else {
    // redirect to containers overview, if no containers are available
    router.push({ name: RouteName.Containers })
  }
})

const container = computed((): Container | undefined => {
  return containers.value.find((c) => c.id === router.currentRoute.value.params.id) || undefined
})
</script>

<style lang="scss" scoped>
.container {
  height: 100%;

  .n-layout-sider {
    background-color: transparent;
  }

  .content {
    display: flex;
    height: 100%;
    flex-flow: column;

    .header {
      display: flex;
      flex-flow: row;
      justify-content: space-between;
      align-items: center;
      padding-bottom: 15px;

      .avatar {
        padding-right: 10px;
      }

      .title {
        font-weight: bold;
        font-size: 1.4em;
        white-space: nowrap;
        text-overflow: ellipsis;
        overflow: hidden;
        min-width: 100px;
        padding-right: 10px;
      }

      .extra {
        margin-left: auto;
      }
    }

    .body {
      flex: 1;
    }

    .footer {
      //
    }
  }
}
</style>
