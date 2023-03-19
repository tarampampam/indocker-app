<template>
  <n-tabs
    default-value="containers"
    justify-content="space-around"
    type="line"
    :value="currentRouteName"
  >
    <n-tab
      v-for="route in this.$router.getRoutes().filter((r) => r.meta.visible)"
      v-on:click="this.$router.push({ name: route.name })"
      :key="route.name"
      :name="route.name"
      style="padding: 2em 1em 1em 1em"
    >
      <n-icon
        v-if="route.meta.icon"
        :component="route.meta.icon"
        :size="18"
        style="padding-right: 0.3em"
      />
      {{ route.meta.title }}
    </n-tab>
  </n-tabs>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { NIcon, NTab, NTabs } from 'naive-ui'

export default defineComponent({
  components: {
    NTabs,
    NTab,
    NIcon
  },
  computed: {
    currentRouteName(): string | undefined {
      if (typeof this.$route.name === 'string') {
        console.log(this.$route.name.split('.')[0])

        return this.$route.name.split('.')[0]
      }

      return undefined
    }
  }
})
</script>

<style lang="scss" scoped></style>
