import hljs from 'highlight.js/lib/core'
import accesslog from 'highlight.js/lib/languages/accesslog'
import { createApp } from 'vue'
import App from '@/App.vue'
import '@/assets/main.css'
import { router } from '@/router'
import { API, APIKey } from '@/api/api'
import { DB, DBKey } from '@/db/db'
import { EmitterKey } from '@/events'
import { DockerStateChannelName } from '@/workers/shared'
import type { AppEvents } from '@/events'
import mitt from 'mitt'

// register the highlight.js languages
hljs.registerLanguage('accesslog', accesslog)

// register the service worker, if it's supported
if ('serviceWorker' in navigator) {
  navigator.serviceWorker
    .register('/service-worker.' + (import.meta.env.DEV ? 'ts' : 'js'), {
      scope: '/',
      type: 'module',
    })
    .catch(console.error)
}

// create the event emitter instance
const emitter = mitt<AppEvents>()

// create the shared worker
new SharedWorker(new URL('@/workers/shared.js', import.meta.url), {
  type: 'module',
  name: 'shared',
}).addEventListener('error', console.error)

// listen to the shared worker messages
new BroadcastChannel(DockerStateChannelName)
  // notify the app that the docker state has been updated (notification was sent from the shared worker)
  .addEventListener('message', () => emitter.emit('dockerStateUpdated'))

// create the VUE app
createApp(App)
  .use(router())
  .use((app) => app.provide(APIKey, new API()))
  .use((app) => app.provide(EmitterKey, emitter))
  .use((app) => app.provide(DBKey, new DB()))
  .mount('#app')
