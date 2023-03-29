<template>
  <n-layout class="container" has-sider>
    <n-layout-sider>
      <containers-list v-if="containers.length" :containers="containers" />
      <containers-list-empty v-else />
    </n-layout-sider>
    <n-layout-content content-style="padding-left: 30px">
      <div v-if="containers.length && container" class="content">
        <div class="header">
          <div class="avatar">
            <n-icon-wrapper :size="32" :border-radius="10" :color="stringToColor(container.id)">
              <n-icon :size="18" :component="DockerIcon" />
            </n-icon-wrapper>
          </div>
          <div class="title" :title="container.name.length > 25 ? container.name : undefined">
            {{ container.name }}
          </div>
          <div class="extra">
            <content-switcher />
          </div>
        </div>
        <div class="body">
          <router-view />
        </div>
        <div class="footer">
          <n-breadcrumb>
            <n-breadcrumb-item :clickable="false">
              Status: Up 57 minutes (healthy)
            </n-breadcrumb-item>
            <n-breadcrumb-item :clickable="false"> Created: 19 March 2023</n-breadcrumb-item>
            <n-breadcrumb-item :clickable="false"> Project: App</n-breadcrumb-item>
          </n-breadcrumb>
        </div>
      </div>
      <n-layout v-else>
        <n-skeleton text :repeat="3" />
        <n-skeleton text :repeat="1" width="66%" />
      </n-layout>
    </n-layout-content>
  </n-layout>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
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
  NSkeleton
} from 'naive-ui'
import { LogoDocker as DockerIcon } from '@vicons/ionicons5'
import { RouteName, goto, id } from '@/router'
import { useRouter } from 'vue-router';

export interface Container {
  id: string
  name: string
  tags: string[]
}

export default defineComponent({
  components: {
    NLayout,
    NLayoutSider,
    NLayoutContent,
    NSkeleton,
    ContainersList,
    ContainersListEmpty,
    ContentSwitcher,
    NBreadcrumb,
    NBreadcrumbItem,
    NIcon,
    NIconWrapper
  },
  setup() {
    return {
      DockerIcon
    }
  },
  mounted(): void {
    const router = useRouter()

    if (this.containers && this.containers.length) {
      const currentID = id()

      // redirect to first container, if no container is selected
      if (!currentID) {
        goto(router, RouteName.ContainerLogs, { id: this.containers[0].id })

        return
      }

      // redirect to first container, if selected container does not exist
      if (!this.containers.find((c) => c.id === currentID)) {
        goto(router, RouteName.ContainerLogs, { id: this.containers[0].id })

        return
      }
    } else {
      // redirect to containers overview, if no containers are available
      goto(router, RouteName.Containers)
    }
  },
  data(): {
    containers: Container[]
  } {
    return {
      containers: [
        {
          id: 'id-1',
          name: 'app-app-1',
          tags: ['docker-compose']
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
            'bar'
          ]
        }
      ]
    }
  },
  methods: {
    stringToColor(str: string): string {
      return stc(str)
    }
  },
  computed: {
    container(): Container | undefined {
      const currentID = id()

      if (currentID) {
        return this.containers.find((c) => c.id === currentID)
      }

      return undefined
    }
  }
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
