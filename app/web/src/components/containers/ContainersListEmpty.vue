<template>
  <NEmpty description="No containers found">
    <template #icon>
      <NIcon>
        <NoContainersIcon />
      </NIcon>
    </template>
  </NEmpty>
</template>

<script setup lang="ts">
import { NEmpty, NIcon } from 'naive-ui'
import { AppsSharp as NoContainersIcon } from '@vicons/ionicons5'
</script>

<style lang="scss" scoped></style>
