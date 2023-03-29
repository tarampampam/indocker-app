<template>
  <n-tabs
    default-value="containers"
    justify-content="space-around"
    type="segment"
    :value="current() === RouteName.ContainerLogs ? 'logs' : 'stats'"
  >
    <n-tab
      name="logs"
      style="padding: 0.3em 3em"
      @click="goto(useRouter(), RouteName.ContainerLogs, { id: id() })"
    >
      <n-icon class="icon" :size="18">
        <logs-icon />
      </n-icon>
      Logs
    </n-tab>
    <n-tab
      name="stats"
      style="padding: 0.3em 3em"
      @click="goto(useRouter(), RouteName.ContainerStats, { id: id() })"
    >
      <n-icon class="icon" :size="18">
        <stats-icon />
      </n-icon>
      Stats
    </n-tab>
  </n-tabs>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { NIcon, NTab, NTabs } from 'naive-ui'
import { BarChart as StatsIcon, ChatboxEllipsesOutline as LogsIcon } from '@vicons/ionicons5'
import { RouteName, goto, id, current } from '@/router'
import { useRouter } from 'vue-router';

export default defineComponent({
  methods: {useRouter},
  setup() {
    return {
      RouteName,
      current,
      goto,
      id
    }
  },
  components: {
    NIcon,
    NTabs,
    NTab,
    LogsIcon,
    StatsIcon
  }
})
</script>

<style lang="scss" scoped>
.icon {
  padding-right: 0.4em;
}
</style>
