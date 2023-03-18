<template>
  <n-icon size="40">
    <docker-icon />
  </n-icon>
  Containers
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { NIcon } from 'naive-ui'
import { LogoDocker as DockerIcon } from '@vicons/ionicons5'

export default defineComponent({
  components: {
    NIcon,
    DockerIcon
  }
})
</script>

<style lang="scss" scoped></style>
