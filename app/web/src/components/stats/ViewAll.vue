<template>
  <NIcon size="40">
    <StatsIcon/>
  </NIcon>
  Stats monitor

  <Line
    ref="chart"
    :options="options"
    :data="data"
  />
</template>

<script setup lang="ts">
import { NIcon } from 'naive-ui'
import { StatsChart as StatsIcon } from '@vicons/ionicons5'
import { Line } from 'vue-chartjs'
import { nextTick, onMounted, ref } from 'vue'
import type { ChartData } from 'chart.js'
import { CategoryScale, Chart as ChartJS, LinearScale, LineElement, PointElement } from 'chart.js'
import type { DefaultDataPoint } from 'chart.js/dist/types'
import { useEmitter } from '@/events'
import { useDB } from '@/db/db'

ChartJS.register(LineElement, CategoryScale, LinearScale, PointElement)

const chart = ref<{ chart: ChartJS } | null>(null)

const options = ref({
  responsive: true,
})

const data = ref<ChartData<'line', DefaultDataPoint<'line'>, string>>({
  labels: ['foo', 'foo', 'foo', 'foo', 'foo', 'foo', 'foo', 'foo'],
  datasets: [{
    label: 'My First Dataset',
    data: [65, 59, 80, 81, null, 55, 40, 123],
    fill: false,
    borderColor: 'rgb(75, 192, 192)',
    tension: 0.1,
  }, {
    label: 'My First Dataset 12313123',
    data: [165, 159, 180, 181, 156, 155, 140],
    fill: false,
    borderColor: 'rgb(75, 19, 192)',
    tension: 0.1,
  }],
})

const db = useDB()
const emitter = useEmitter()

onMounted(async () => {
  // (await db.getContainersMemoryUsage()).map((item) => {
  //   data.value.labels?.push(`${
  //     ("0" + item.ts.getHours()).slice(-2)
  //   }:${
  //     ("0" + item.ts.getMinutes()).slice(-2)
  //   }:${
  //     ("0" + item.ts.getSeconds()).slice(-2)
  //   }`)
  //
  //   data.value.datasets?.forEach((dataset) => {
  //     // dataset.data?.push(item.)
  //   })
  // })
  //
  // const usage = await db.getContainersMemoryUsage()
  //
  // data.value.labels?.push(...usage.map((item) => `${
  //   ("0" + item.ts.getHours()).slice(-2)
  // }:${
  //   ("0" + item.ts.getMinutes()).slice(-2)
  // }:${
  //   ("0" + item.ts.getSeconds()).slice(-2)
  // }`))
  //
  // console.log('data is ready')
  //
  // if (chart.value) {
  //   await nextTick(() => {
  //     console.log('updating chart')
  //     chart.value?.chart.update()
  //   })
  // }
  // chart.value?.chart.update()
  emitter.on('dockerStateUpdated', async () => {


    // console.debug('event received', map)
    // console.debug('timestamps', timestamps)
  })
})
</script>

<style lang="scss" scoped></style>
