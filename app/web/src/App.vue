<template>
  <n-config-provider :theme="theme" :hljs="hljs">
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
import hljs from 'highlight.js/lib/core'
import accesslog from 'highlight.js/lib/languages/accesslog'
import { defineComponent, ref } from 'vue'
import PageFooter from './components/PageFooter.vue'
import TopNavigation from './components/TopNavigation.vue'
import { darkTheme, lightTheme, NCard, NConfigProvider, NGlobalStyle } from 'naive-ui'

hljs.registerLanguage('accesslog', accesslog)

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
      hljs,
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
