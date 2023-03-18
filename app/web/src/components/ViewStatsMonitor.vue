<template>
  <n-icon size="40">
    <stats-icon />
  </n-icon>
  Stats monitor
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { NIcon } from 'naive-ui'
import { StatsChart as StatsIcon } from '@vicons/ionicons5'

export default defineComponent({
  components: {
    NIcon,
    StatsIcon
  }
})
</script>

<style lang="scss" scoped></style>
