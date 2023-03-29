<template>
  <NTabs
    default-value="containers"
    justify-content="space-around"
    type="line"
    :value="current(true)"
  >
    <NTab
      v-for="route in visible()"
      @click="goto(this.$router, route.name)"
      :key="route.name"
      :name="route.name"
      style="padding: 2em 1em 1em 1em"
    >
      <NIcon
        v-if="route.meta.icon"
        :component="route.meta.icon"
        :size="18"
        style="padding-right: 0.3em"
      />
      {{ route.meta.title }}
    </NTab>
  </NTabs>
</template>

<script setup lang="ts">
import { NIcon, NTab, NTabs } from 'naive-ui'
import { current, goto, visible } from '@/router'
</script>

<style lang="scss" scoped></style>
