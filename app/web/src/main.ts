import hljs from 'highlight.js/lib/core'
import accesslog from 'highlight.js/lib/languages/accesslog'
import { createApp } from 'vue'
import App from '@/App.vue'
import './assets/main.css'
import { router } from './router'
import API from '@/api/api';

hljs.registerLanguage('accesslog', accesslog)

createApp(App).use(router()).mount('#app')

const api = new API()

api.version().then((version) => {
  console.log(version)
})

api.latestVersion().then((latest) => {
  console.log(latest)
})
