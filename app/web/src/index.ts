import {createApp} from 'vue'
import mitt from 'mitt'
import App from './views/app.vue'

createApp(App)
  .use((app) => app.config.globalProperties.$emitter = mitt())
  .mount('#app')
