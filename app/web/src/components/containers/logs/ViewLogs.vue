<template>
  <n-log ref="logInst" :log="log.join('\n')" language="accesslog" trim />
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { NLog } from 'naive-ui'

export default defineComponent({
  components: {
    NLog,
  },
  data(): {
    log: string[]
  } {
    return {
      log: ['foo', 'bar', '123456.7890'],
    }
  }
})
</script>

<style lang="scss" scoped></style>
