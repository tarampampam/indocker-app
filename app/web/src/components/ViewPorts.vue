<template>
  <n-icon size="40">
    <ports-icon />
  </n-icon>
  Ports
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { NIcon } from 'naive-ui'
import { SwapVertical as PortsIcon } from '@vicons/ionicons5'

export default defineComponent({
  components: {
    NIcon,
    PortsIcon
  }
})
</script>

<style lang="scss" scoped></style>
