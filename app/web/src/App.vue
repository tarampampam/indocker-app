<template>
  <n-config-provider :theme="theme">
    <n-tabs
      default-value="containers"
      justify-content="space-around"
      type="line"
      :value="this.$route.name"
    >
      <n-tab
        v-for="route in this.$router.getRoutes().filter((r) => r.meta.visible)"
        v-on:click="this.$router.push({ name: route.name })"
        :key="route.name"
        :name="route.name"
        :style="{padding: '2em 0 1em 0'}"
      >
        {{ route.meta.title }}
      </n-tab>
    </n-tabs>
    <main>
      <router-view></router-view>
    </main>
    <n-card class="footer">
      <template #header>
        <page-footer/>
      </template>
    </n-card>

    <n-global-style/>
  </n-config-provider>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue'
import { darkTheme, lightTheme, NCard, NConfigProvider, NGlobalStyle, NTab, NTabs } from 'naive-ui'
import PageFooter from './components/PageFooter.vue'

export default defineComponent({
  components: {
    NConfigProvider,
    NGlobalStyle,
    NTabs,
    NCard,
    NTab,
    PageFooter,
  },
  setup() {
    const dark = darkTheme
    const light = lightTheme
    const mediaSelector = '(prefers-color-scheme: dark)'
    const theme = ref(light)

    if (window.matchMedia) {
      theme.value = window.matchMedia(mediaSelector).matches ? dark : light

      window.matchMedia(mediaSelector).addEventListener('change', (event) => {
        theme.value = event.matches ? dark : light
      })
    }

    return {
      theme: theme,
    }
  },
})
</script>

<style lang="scss" scoped>
.n-config-provider {
  height: 100vh;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;

  main {
    flex: 1;
    box-sizing: border-box;
    padding: 20px 30px;
  }

  .footer {
    border: none;
    border-radius: 0;
  }
}
</style>
