<template>
  <n-layout>
    <n-layout has-sider>
      <n-layout-sider>
        <n-list v-if="containers.length" :show-divider="false" hoverable clickable>
          <n-list-item
            v-for="container in containers"
            :key="container.id"
            v-on:click="
              this.$router.push({ name: 'containers.logs', params: { id: container.id } })
            "
            :class="{ active: this.$route.params.id === container.id }"
          >
            <template #prefix>
              <n-icon-wrapper :size="14" :border-radius="14" :color="theme.defaultColor" />
            </template>
            <template #default>
              <n-thing :title="container.name">
                <template #description>
                  <n-space>
                    <n-tag v-for="tag in container.tags" :key="tag" type="info" size="small">
                      {{ tag }}
                    </n-tag>
                  </n-space>
                </template>
              </n-thing>
            </template>
            <template #suffix>
              <n-icon :size="25" :depth="4">
                <arrow-icon />
              </n-icon>
            </template>
          </n-list-item>
        </n-list>
        <n-empty v-else description="No containers found">
          <template #icon>
            <n-icon>
              <no-containers-icon />
            </n-icon>
          </template>
        </n-empty>
      </n-layout-sider>
      <n-layout-content content-style="padding: 0 1em;">
        <n-layout v-if="containers.length">
          <n-space justify="space-between">
            <n-thing>
              <!-- anything -->
            </n-thing>
            <n-tabs
              default-value="containers"
              justify-content="space-around"
              type="segment"
              :value="this.$router.currentRoute.value.name === 'containers.logs' ? 'logs' : 'stats'"
            >
              <n-tab
                name="logs"
                style="padding: 0.3em 3em"
                v-on:click="
                  this.$router.push({
                    name: 'containers.logs',
                    params: { id: this.$router.currentRoute.value.params.id }
                  })
                "
              >
                <n-icon :size="18" style="padding-right: 0.4em">
                  <logs-icon />
                </n-icon>
                Logs
              </n-tab>
              <n-tab
                name="stats"
                style="padding: 0.3em 3em"
                v-on:click="
                  this.$router.push({
                    name: 'containers.stats',
                    params: { id: this.$router.currentRoute.value.params.id }
                  })
                "
              >
                <n-icon :size="18" style="padding-right: 0.4em">
                  <stats-icon />
                </n-icon>
                Stats
              </n-tab>
            </n-tabs>
          </n-space>
          <router-view />
        </n-layout>
        <n-skeleton v-else text :repeat="2" />
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import {
  NEmpty,
  NIcon,
  NIconWrapper,
  NLayout,
  NLayoutContent,
  NLayoutSider,
  NList,
  NListItem,
  NSkeleton,
  NSpace,
  NTab,
  NTabs,
  NTag,
  NThing,
  useThemeVars
} from 'naive-ui'
import {
  AppsSharp as NoContainersIcon,
  ArrowForwardCircleOutline as ArrowIcon,
  ChatboxEllipsesOutline as LogsIcon,
  BarChart as StatsIcon
} from '@vicons/ionicons5'

export default defineComponent({
  components: {
    NLayout,
    NLayoutSider,
    NLayoutContent,
    NIconWrapper,
    NListItem,
    NThing,
    NSpace,
    NList,
    NIcon,
    NTag,
    NEmpty,
    NSkeleton,
    ArrowIcon,
    NTabs,
    NTab,
    LogsIcon,
    StatsIcon,
    NoContainersIcon
  },
  mounted() {
    if (this.containers && this.containers.length) {
      // redirect to first container, if no container is selected
      if (!this.$route.params.id) {
        this.$router.push({ name: 'containers.logs', params: { id: this.containers[0].id } })

        return
      }

      // redirect to first container, if selected container does not exist
      if (!this.containers.find((c) => c.id === this.$route.params.id)) {
        this.$router.push({ name: 'containers.logs', params: { id: this.containers[0].id } })

        return
      }
    } else {
      // redirect to containers overview, if no containers are available
      this.$router.push({ name: 'containers' })
    }
  },
  data(): {
    theme: {
      activeColor: string
      successColor: string
      warningColor: string
      errorColor: string
      defaultColor: string
    }
    containers: { id: string; name: string; tags: Array<string> }[]
  } {
    const theme = useThemeVars()

    return {
      theme: {
        activeColor: theme.value.infoColor,
        successColor: theme.value.successColor,
        warningColor: theme.value.warningColor,
        errorColor: theme.value.errorColor,
        defaultColor: theme.value.iconColorDisabled
      },
      containers: [
        {
          id: 'id-1',
          name: 'app-app-1',
          tags: ['docker-compose']
        },
        {
          id: 'id-2',
          name: 'app-app-2',
          tags: ['foo', 'bar']
        }
      ]
    }
  }
})
</script>

<style lang="scss" scoped>
.active {
  box-shadow: -2px 0 v-bind('theme.activeColor') inset;
}
</style>
