<template>
  <div class="content">
    <div class="hack">
      <NLog :log="logLines" ref="logContainer" language="accesslog" style="height: 100%" trim />

      <div class="controls">
        <NButtonGroup size="tiny">
          <NButton :focusable="false" @click="clear" secondary round>
            <template #icon>
              <NIcon>
                <CleanIcon />
              </NIcon>
            </template>
            Clean logs
          </NButton>
          <NButton v-if="!follow" @click="enableFollowing" :focusable="false" secondary round>
            <template #icon>
              <NIcon>
                <FollowIcon />
              </NIcon>
            </template>
            Follow the logs
          </NButton>
          <NButton v-else @click="disableFollowing" :focusable="false" secondary round>
            <template #icon>
              <NIcon>
                <StopIcon />
              </NIcon>
            </template>
            Disable following
          </NButton>
        </NButtonGroup>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'
import type { LogInst } from 'naive-ui'
import { NButton, NButtonGroup, NIcon, NLog } from 'naive-ui'
import { ArrowDown as FollowIcon, Stop as StopIcon, TrashBin as CleanIcon } from '@vicons/ionicons5'

const logContainer = ref<LogInst | null>(null) // ref to the log component
const logLines = ref<string>('') // log lines as a string, for rendering
const lines = ref<string[]>([]) // log lines as an array (source of truth)
const follow = ref<boolean>(true) // whether to follow the logs

/** Enable following the logs */
function enableFollowing(): void {
  follow.value = true
  scrollToBottom()
}

/** Disable following the logs */
function disableFollowing(): void {
  follow.value = false
}

/** Clear the logs */
function clear(): void {
  lines.value = []
}

/** Scroll to the bottom of the log */
function scrollToBottom(): void {
  if (logContainer.value) {
    nextTick(() => {
      logContainer.value?.scrollTo({ position: 'bottom', slient: false })
    })
  }
}

/** Watch for changes in the log lines */
watch(
  lines,
  (): void => {
    logLines.value = lines.value.join('\n')

    if (follow.value) {
      scrollToBottom()
    }
  },
  { deep: true }
)

onMounted((): void => {
  for (let i = 0; i < 50; i++) {
    // TODO: just for a demo
    lines.value.push(
      `20.164.151.111 - - [20/Aug/2015:22:20:18 -0400] "GET /mywebpage/index.php HTTP/1.1" 403 772 "-" ` +
        `"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_6_8) AppleWebKit/535.1 (KHTML, like Gecko) ` +
        `Chrome/13.0.782.220 Safari/535.1"`
    )
    lines.value.push(`{"message": "foo bar", "ts": 123456, "bool": true}`)
  }

  window.setInterval(() => {
    // TODO: just for demo
    lines.value.push('new line. rnd = ' + Math.random())
  }, 1000)

  scrollToBottom()
})
</script>

<style lang="scss" scoped>
.content {
  position: relative;
  height: 100%;

  .hack {
    position: absolute;
    left: 0;
    right: 0;
    top: 0;
    bottom: 0;

    .controls {
      position: absolute;
      right: 15px;
      bottom: 7px;
    }
  }
}
</style>
