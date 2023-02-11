<template>
  <main>
    Hello, world!
  </main>
</template>

<script lang="ts">
import {defineComponent} from 'vue'

export default defineComponent({
  components: {},

  data(): {[key: string]: unknown} {
    return {}
  },

  created(): void {
    // ...
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
