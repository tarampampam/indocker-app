<template>
  <NTabs
    default-value="containers"
    justify-content="space-around"
    type="segment"
    :value="current() === RouteName.ContainerLogs ? 'logs' : 'stats'"
  >
    <NTab
      name="logs"
      style="padding: 0.3em 3em"
      @click="goto(useRouter(), RouteName.ContainerLogs, { id: id() })"
    >
      <NIcon class="icon" :size="18">
        <LogsIcon />
      </NIcon>
      Logs
    </NTab>
    <NTab
      name="stats"
      style="padding: 0.3em 3em"
      @click="goto(useRouter(), RouteName.ContainerStats, { id: id() })"
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
import { RouteName, goto, id, current } from '@/router'
import { useRouter } from 'vue-router'
</script>

<style lang="scss" scoped>
.icon {
  padding-right: 0.4em;
}
</style>
