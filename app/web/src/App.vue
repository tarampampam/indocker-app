<template>
  <div v-if="isUnderConstruction" style="display: flex; height: 100vh; align-items: center; justify-content: center">
    Monitor is under construction
  </div>
  <NConfigProvider v-else :theme="theme" :hljs="hljs">
    <TopNavigation />
    <main>
      <router-view></router-view>
    </main>
    <NCard class="footer">
      <template #header>
        <page-footer />
      </template>
    </NCard>

    <NGlobalStyle />
  </NConfigProvider>
</template>

<script setup lang="ts">
import hljs from 'highlight.js/lib/core'
import { ref, onMounted, onBeforeMount } from 'vue'
import PageFooter from '@/components/PageFooter.vue'
import TopNavigation from '@/components/TopNavigation.vue'
import { darkTheme, lightTheme, NCard, NConfigProvider, NGlobalStyle } from 'naive-ui'

const theme = ref(lightTheme)

const isUnderConstruction = true // set false to enable

onBeforeMount((): void => {
  if (window.matchMedia) {
    // is media query supported?
    const mediaSelector = '(prefers-color-scheme: dark)'

    theme.value = window.matchMedia(mediaSelector).matches ? darkTheme : lightTheme

    window.matchMedia(mediaSelector).addEventListener('change', (event): void => {
      theme.value = event.matches ? darkTheme : lightTheme
    })
  }
})

onMounted((): void => {
  //
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
