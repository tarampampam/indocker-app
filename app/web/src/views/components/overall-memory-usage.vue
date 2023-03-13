<template>
  <Line
    ref="chart"
    :options="chartOptions"
    :data="chartData"
  />
  <button v-on:click="test">test</button>
</template>

<script lang="ts">
import {Line} from 'vue-chartjs'
import {defineComponent} from 'vue'
import {Chart, ChartOptions} from 'chart.js'
import {ChartData} from 'chart.js/dist/types'
import stc from 'string-to-color'
import {EventNames, MemoryUsageUpdatedEvent} from '../events'

export default defineComponent({
  components: {
    Line,
  },

  props: {
    size: {
      type: Number,
      default: 60,
    },
  },

  data(): {
    chartOptions: ChartOptions
    chartData: ChartData
    [key: string]: unknown
  } {
    return {
      chartOptions: {
        // responsive: true,
        mode: 'nearest',
        // maintainAspectRatio: false,
      } as ChartOptions,
      chartData: {
        labels: [],
        datasets: [],
      } as ChartData,
    }
  },

  mounted() {
    this.$emitter.on(EventNames.MemoryUsageUpdated, this.onUpdate) // subscribe to global event
  },

  unmounted() {
    this.$emitter.off(EventNames.MemoryUsageUpdated, this.onUpdate) // and unsubscribe
  },

  methods: {
    /** Format unix timestamp to human-readable time (HH:MM:SS) */
    formatTimestamp(ts: number): string {
      const d = new Date(ts * 1000)

      return [
        d.getHours().toString().padStart(2, '0'),
        d.getMinutes().toString().padStart(2, '0'),
        d.getSeconds().toString().padStart(2, '0'),
      ].join(':')
    },

    /** Convert string to color */
    stringToColor(str: string): string {
      return stc(str)
    },

    /** Get chart instance */
    chart(): Chart {
      return (this.$refs.chart as any).chart // https://vue-chartjs.org/guide/#access-to-chart-instance
    },

    /** Update chart data */
    onUpdate(event: MemoryUsageUpdatedEvent): void {
      console.debug('onUpdate', event)

      // this.chart().data.labels?.push('dfsf')
      this.chart().update()
    },

    test(): void {
      console.warn('test', this.chart())
    }
  },

  computed: {
    // chartData(): ChartData {
    //   const size = 60 // seconds count
    //   const now = Math.round(Date.now() / 1000)
    //   const columns = new Array(size)
    //     .fill(0)
    //     .map((_, i) => now - (size - 1) + i) // set chart column names (unix timestamp
    //
    //   const containerIDs = Object.keys(this.containers)
    //
    //   return {
    //     labels: columns.map(c => { // format unix timestamp to human-readable time (HH:MM:SS)
    //       const d = new Date(c * 1000)
    //
    //       return [
    //         d.getHours().toString().padStart(2, '0'),
    //         d.getMinutes().toString().padStart(2, '0'),
    //         d.getSeconds().toString().padStart(2, '0'),
    //       ].join(':')
    //     }),
    //     datasets: new Array(containerIDs.length)
    //       .fill({} as ChartDataset) // init with empty objects
    //       .map((_, i) => {
    //         const containerID = containerIDs[i]
    //         const container = this.containers[containerID]
    //         const stats = container.stats.slice(-size) // get last N stats
    //
    //         const data: (number | null)[] = new Array(size).fill(null) // init with null values
    //
    //         stats.forEach((stat) => {
    //           const ts = Math.round(Date.parse(stat.read) / 1000)
    //           const allowedDelta = 1 // seconds
    //
    //           const closest = columns.reduce((prev, curr) => Math.abs(curr - ts) < Math.abs(prev - ts) ? curr : prev)
    //
    //           if (ts + allowedDelta < closest || ts - allowedDelta > closest)  {
    //             return // skip if timestamp is not in allowed delta
    //           }
    //
    //           data[columns.indexOf(closest)] = stat.memory_stats.usage / 1024 / 1024 // convert bytes to megabytes
    //         })
    //
    //         return {
    //           label: container.inspect?.Name?.slice(1) || containerID,
    //           backgroundColor: stc(containerIDs[i]),
    //           data: data,
    //         }
    //       }),
    //   }
    // }
  }
})
</script>

<style lang="scss" scoped>
</style>
