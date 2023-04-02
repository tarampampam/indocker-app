<template>
  <NTabs
    default-value="containers"
    justify-content="space-around"
    type="segment"
    :value="router.currentRoute.value.name === RouteName.ContainerLogs ? 'logs' : 'stats'"
  >
    <NTab
      name="logs"
      style="padding: 0.3em 3em"
      @click="
        router.push({
          name: RouteName.ContainerLogs,
          params: { id: router.currentRoute.value.params.id },
        })
      "
    >
      <NIcon class="icon" :size="18">
        <LogsIcon />
      </NIcon>
      Logs
    </NTab>

    <NTab
      name="stats"
      style="padding: 0.3em 3em"
      @click="
        router.push({
          name: RouteName.ContainerStats,
          params: { id: router.currentRoute.value.params.id },
        })
      "
    >
      <NIcon class="icon" :size="18">
        <StatsIcon />
      </NIcon>
      Stats
    </NTab>
  </NTabs>
</template>

<script setup lang="ts">
import { NIcon, NTab, NTabs } from 'naive-ui'
import { BarChart as StatsIcon, ChatboxEllipsesOutline as LogsIcon } from '@vicons/ionicons5'
import { RouteName } from '@/router'
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<style lang="scss" scoped>
.icon {
  padding-right: 0.4em;
}
</style>
