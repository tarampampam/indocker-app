<template>
  <div class="content">
    <div class="hack">
      <n-log :log="lines" ref="log" language="accesslog" style="height: 100%" trim />

      <div class="controls">
        <n-button-group size="tiny">
          <n-button :focusable="false" @click="lines = ''" secondary round>
            <template #icon>
              <n-icon>
                <clean-icon />
              </n-icon>
            </template>
            Clean logs
          </n-button>
          <n-button
            v-if="!follow"
            @click="
              () => {
                follow = true
                scrollToBottom()
              }
            "
            :focusable="false"
            secondary
            round
          >
            <template #icon>
              <n-icon>
                <follow-icon />
              </n-icon>
            </template>
            Follow the logs
          </n-button>
          <n-button v-else @click="follow = false" :focusable="false" secondary round>
            <template #icon>
              <n-icon>
                <stop-icon />
              </n-icon>
            </template>
            Disable following
          </n-button>
        </n-button-group>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { nextTick, defineComponent } from 'vue'
import { NButton, NButtonGroup, NLog, NIcon } from 'naive-ui'
import type { LogInst } from 'naive-ui'
import { Stop as StopIcon, ArrowDown as FollowIcon, TrashBin as CleanIcon } from '@vicons/ionicons5'

export default defineComponent({
  components: {
    NLog,
    NIcon,
    NButton,
    NButtonGroup,
    StopIcon,
    FollowIcon,
    CleanIcon
  },
  data(): {
    lines: string
    follow: boolean
  } {
    const lines: string[] = []

    for (let i = 0; i < 50; i++) {
      lines.push(
        `20.164.151.111 - - [20/Aug/2015:22:20:18 -0400] "GET /mywebpage/index.php HTTP/1.1" 403 772 "-" ` +
          `"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_6_8) AppleWebKit/535.1 (KHTML, like Gecko) ` +
          `Chrome/13.0.782.220 Safari/535.1"`
      )
      lines.push(`{"message": "foo bar", "ts": 123456, "bool": true}`)
    }

    return {
      lines: lines.join('\n'),
      follow: true
    }
  },
  mounted(): void {
    window.setInterval(() => {
      // TODO: just for demo
      this.lines += 'new line. rnd = ' + Math.random() + '\n'
    }, 1000)

    if (this.follow) {
      this.scrollToBottom()
    }
  },

  watch: {
    lines(): void {
      if (this.follow) {
        this.scrollToBottom()
      }
    }
  },

  methods: {
    scrollToBottom(): void {
      const log = this.$refs.log as LogInst | undefined

      if (log) {
        nextTick(() => {
          log.scrollTo({ position: 'bottom', slient: false })
        })
      }
    }
  }
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
