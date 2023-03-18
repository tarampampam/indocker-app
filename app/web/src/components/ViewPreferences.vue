<template>
  <n-icon size="40">
    <preferences-icon />
  </n-icon>
  Stats monitor
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { NIcon } from 'naive-ui'
import { Build as PreferencesIcon } from '@vicons/ionicons5'

export default defineComponent({
  components: {
    NIcon,
    PreferencesIcon
  }
})
</script>

<style lang="scss" scoped></style>
