import hljs from 'highlight.js/lib/core'
import accesslog from 'highlight.js/lib/languages/accesslog'
import { createApp } from 'vue'
import App from '@/App.vue'
import './assets/main.css'
import { router } from './router'
import { discover } from '@/api/api';

hljs.registerLanguage('accesslog', accesslog)

createApp(App).use(router()).mount('#app')

discover().then(console.log)
