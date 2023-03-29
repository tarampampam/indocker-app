<template>
  <NTabs default-value="containers" justify-content="space-around" type="line" :value="current">
    <NTab
      v-for="route in visible"
      @click="router.push({ name: route.name })"
      :key="route.name"
      :name="route.name.toString()"
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
import { RouteName } from '@/router'
import { useRouter } from 'vue-router'
import type { RouteRecord } from 'vue-router'
import { computed, ref } from 'vue'

const router = useRouter()

const visible = ref<RouteRecord[]>(router.getRoutes().filter((route) => route.meta.visible))

const current = computed((): RouteName | undefined => {
  const name = useRouter().currentRoute?.value?.name

  if (typeof name === 'string') {
    return name.split('.')[0] as RouteName
  }

  return undefined
})
</script>

<style lang="scss" scoped></style>
