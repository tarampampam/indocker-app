import {createApp} from 'vue'
import { Chart as ChartJS, Title, Tooltip, Legend, LineElement, CategoryScale, LinearScale, PointElement } from 'chart.js'
import mitt from 'mitt'
import mainApp from './views/main-app.vue'

ChartJS.register(Title, Tooltip, Legend, LineElement, CategoryScale, LinearScale, PointElement)

createApp(mainApp)
  .use((app) => app.config.globalProperties.$emitter = mitt())
  .mount('#app')
