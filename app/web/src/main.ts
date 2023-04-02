import hljs from 'highlight.js/lib/core'
import accesslog from 'highlight.js/lib/languages/accesslog'
import { createApp } from 'vue'
import App from '@/App.vue'
import '@/assets/main.css'
import { router } from '@/router'
import { API, APIKey } from '@/api/api'
import { DockerStateChannelName } from '@/workers/shared'

hljs.registerLanguage('accesslog', accesslog)

if ('serviceWorker' in navigator) {
  navigator.serviceWorker
    .register('/service-worker.' + (import.meta.env.DEV ? 'ts' : 'js'), {
      scope: '/',
      type: 'module'
    })
    .catch(console.error)
}

new SharedWorker(new URL('@/workers/shared.js', import.meta.url), {
  type: 'module',
  name: 'shared'
}).addEventListener('error', console.error)

new BroadcastChannel(DockerStateChannelName).addEventListener('message', console.info)

createApp(App)
  .use(router())
  .use((app) => app.provide(APIKey, new API()))
  .mount('#app')
