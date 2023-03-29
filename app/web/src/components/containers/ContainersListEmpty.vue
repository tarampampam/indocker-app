<template>
  <n-empty description="No containers found">
    <template #icon>
      <n-icon>
        <no-containers-icon />
      </n-icon>
    </template>
  </n-empty>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { NEmpty, NIcon } from 'naive-ui'
import { AppsSharp as NoContainersIcon } from '@vicons/ionicons5'

export default defineComponent({
  components: {
    NIcon,
    NEmpty,
    NoContainersIcon
  }
})
</script>

<style lang="scss" scoped></style>
