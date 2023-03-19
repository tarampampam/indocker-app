<template>
  <n-config-provider :theme="theme">
    <top-navigation />
    <main>
      <router-view></router-view>
    </main>
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
import PageFooter from './components/PageFooter.vue'
import TopNavigation from './components/TopNavigation.vue'
import { darkTheme, lightTheme, NCard, NConfigProvider, NGlobalStyle } from 'naive-ui'

export default defineComponent({
  components: {
    NConfigProvider,
    NGlobalStyle,
    NCard,
    PageFooter,
    TopNavigation
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
