<template>
  <div class="center">
    <NResult status="404" title="Not Found" size="huge" :description="randomDescription()" />
  </div>
</template>

<script setup lang="ts">
import { NResult } from 'naive-ui'

function randomDescription(): string {
  const descriptions: string[] = [
    'Oops!',
    'Are you lost?',
    `Sorry, we can't find that page!`,
    'Looks like this page got lost in the matrix',
    'Sorry, looks like we sent you the wrong way',
    `You're in the wrong place`,
    `We're not quite sure what went wrong`,
  ]

  return descriptions[Math.floor(Math.random() * descriptions.length)]
}
</script>

<style lang="scss" scoped>
.center {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
