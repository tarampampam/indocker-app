<template>
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
      <n-icon class="icon" :size="18">
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

export default defineComponent({
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
