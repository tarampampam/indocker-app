<template>
  <div class="center">
    <n-result status="404" title="Not Found" size="huge" :description="randomDescription()" />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { NResult } from 'naive-ui'

const descriptions: string[] = [
  'Oops!',
  'Are you lost?',
  `Sorry, we can't find that page!`,
  'Looks like this page got lost in the matrix',
  'Sorry, looks like we sent you the wrong way',
  `You're in the wrong place`,
  `We're not quite sure what went wrong`
]

export default defineComponent({
  components: {
    NResult
  },
  methods: {
    randomDescription(): string {
      return descriptions[Math.floor(Math.random() * descriptions.length)]
    }
  }
})
</script>

<style lang="scss" scoped>
.center {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
