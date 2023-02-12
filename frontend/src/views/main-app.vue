<template>
  <overall-memory-usage/>
</template>

<script lang="ts">
import {defineComponent} from 'vue'
import API from '../api'
import OverallMemoryUsage from './components/overall-memory-usage.vue'
import {EventNames, MemoryUsageUpdatedEvent} from './events';

const errorsHandler = console.error

const api = new API('https://monitor.indocker.app/docker-info/')

export default defineComponent({
  components: {
    'overall-memory-usage': OverallMemoryUsage,
  },

  data(): { [key: string]: unknown } {
    return {
      // ...
    }
  },

  async created() {
    window.setInterval(this.update, 1000)
    this.update()
  },

  methods: {
    update(): void {
      api.containersList()
        .then(containers => {
          const ids = containers.map(c => c.Id)

          Promise
            .all(ids.map(id => api.stats(id)))
            .then(stats => {
              const event: MemoryUsageUpdatedEvent = {
                containers: {},
              }

              stats.forEach((stat, i) => {
                event.containers[ids[i]] = {
                  totalMemoryUsageInBytes: stat.memory_stats.usage,
                }
              })

              this.$emitter.emit(EventNames.MemoryUsageUpdated, event)
            })
            .catch(errorsHandler)
        })
        .catch(errorsHandler)
    },
  },

  computed: {
    // ...
  },
})
</script>

<style lang="scss">
// $web-font-path: false; // disable external font named "Lato"

@import "~bootstrap/scss/bootstrap";
</style>
