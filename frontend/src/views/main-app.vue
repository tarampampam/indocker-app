<template>
  <main>
    Hello, world!
  </main>
</template>

<script lang="ts">
import {defineComponent} from 'vue'
import API from '../api'

const api = new API('https://monitor.indocker.app/docker-info/')

export default defineComponent({
  components: {},

  data(): {[key: string]: unknown} {
    return {}
  },

  async created() {
    await Promise.all([
      api.ping(),
      api.version(),
      api.containersList(),
      api.inspect('f8'),
      api.stats('f8'),
      api.logs('f8', 5),
    ]).then(([
      ping,
      version,
      containersList,
      inspect,
      stats,
      logs,
    ]) => {
      console.log(ping)
      console.log(version)
      console.log(containersList)
      console.log(inspect)
      console.log(stats)
      console.log(logs)
    })
  },

  computed: {
    // ...
  },

  methods: {
    // ...
  },
})
</script>

<style lang="scss">
$web-font-path: false; // disable external font named "Lato"

@import "~bootstrap/scss/bootstrap";
</style>
