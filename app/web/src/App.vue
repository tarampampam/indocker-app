<template>
  <n-config-provider :theme="theme">
    <n-tabs class="main" default-value="containers" justify-content="space-around" type="line">
      <n-tab-pane name="containers" tab="Containers">
        <view-containers />
      </n-tab-pane>
      <n-tab-pane name="stats-monitor" tab="Stats monitor">
        <view-stats-monitor />
      </n-tab-pane>
      <n-tab-pane name="ports" tab="Ports">
        <view-ports />
      </n-tab-pane>
      <n-tab-pane name="preferences" tab="Preferences">
        <view-preferences />
      </n-tab-pane>
    </n-tabs>
    <n-card class="footer">
      <template #header>
        <page-footer />
      </template>
    </n-card>

    <n-global-style />
  </n-config-provider>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue'
import {
  darkTheme,
  lightTheme,
  NCard,
  NConfigProvider,
  NGlobalStyle,
  NTabPane,
  NTabs
} from 'naive-ui'
import ViewContainers from './components/ViewContainers.vue'
import ViewStatsMonitor from './components/ViewStatsMonitor.vue'
import ViewPorts from './components/ViewPorts.vue'
import ViewPreferences from './components/ViewPreferences.vue'
import PageFooter from './components/PageFooter.vue'

export default defineComponent({
  components: {
    NConfigProvider,
    NGlobalStyle,
    NTabPane,
    NTabs,
    NCard,
    ViewContainers,
    ViewStatsMonitor,
    ViewPorts,
    ViewPreferences,
    PageFooter
  },
  setup() {
    const dark = darkTheme,
      light = lightTheme
    const mediaSelector = '(prefers-color-scheme: dark)'
    const theme = ref(light)

    if (window.matchMedia) {
      theme.value = window.matchMedia(mediaSelector).matches ? dark : light

      window.matchMedia(mediaSelector).addEventListener('change', (event) => {
        theme.value = event.matches ? dark : light
      })
    }

    return {
      theme: theme
    }
  }
})
</script>

<style lang="scss" scoped>
.n-config-provider {
  height: 100vh;
  display: flex;
  flex-direction: column;

  .main {
    flex: 1;

    .n-tab-pane {
      box-sizing: border-box;
      padding: 20px 30px;
    }
  }

  .footer {
    border: none;
    border-radius: 0;
  }
}
</style>
