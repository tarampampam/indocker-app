import { init, getInstanceByDom } from 'echarts'
import type { EChartsType, EChartsOption, SetOptionOpts, ECElementEvent } from 'echarts'
import React, { useEffect, useRef, useState } from 'react'
import type { CSSProperties } from 'react'
import { isDarkMode } from '~/theme'

/** @link https://dev.to/manufac/using-apache-echarts-with-react-and-typescript-353k */
export default function Chart({
  option,
  style,
  settings,
  loading,
  events,
}: {
  option: EChartsOption
  style?: CSSProperties
  settings?: SetOptionOpts
  loading?: boolean
  events?: Record<Parameters<EChartsType['on']>[0], (event: ECElementEvent) => void>
}): React.JSX.Element {
  const [isDark, disableThemeChangesWatch] = isDarkMode((isDark) => setTheme?.(isDark ? 'dark' : 'light'))
  const [theme, setTheme] = useState<'light' | 'dark'>(isDark ? 'dark' : 'light')
  const chartRef = useRef<HTMLDivElement>(null)

  // stop the dark mode preference listener when the component is unmounted
  useEffect(() => disableThemeChangesWatch, []) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    let chart: EChartsType | undefined

    if (chartRef.current !== null) {
      chart = init(chartRef.current, theme) // initialize chart

      if (events) {
        for (const [event, handler] of Object.entries(events)) {
          chart.on(event, (e) => handler(e as ECElementEvent)) // add event listeners
        }
      }
    }

    const resizeChart = (): void => chart?.resize() // add chart resize listener
    window.addEventListener('resize', resizeChart, { passive: true })

    // Return cleanup function
    return (): void => {
      chart?.dispose()

      window.removeEventListener('resize', resizeChart)
    }
  }, [theme, events])

  useEffect(() => {
    if (chartRef.current !== null) {
      getInstanceByDom(chartRef.current)?.setOption(option, settings) // update chart
    }
  }, [option, settings, theme, events])

  useEffect(() => {
    if (chartRef.current !== null) {
      const chart = getInstanceByDom(chartRef.current)

      if (chart) {
        loading ? chart.showLoading() : chart.hideLoading()
      }
    }
  }, [loading, theme])

  return <div ref={chartRef} style={{ width: '100%', height: '100px', ...style }} />
}
